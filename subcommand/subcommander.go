package subcommand

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/gogolfing/cli"
)

//SubCommander provides registering multiple SubCommands and executing them from
//command line arguments.
//
//Note that SubCommander is NOT safe for use with multiple goroutines.
type SubCommander struct {
	//CommandName is used in error and help output. It should be the name of the
	//program that was invoked.
	CommandName string

	//GlobalFlags is a FlagSetter that is used for setting global flags for subcommands.
	//
	//This value may be called multiple times with different values for f if argument
	//parsing fails. This is done to obtain possible help and error output.
	//If there are no errors during the argument parsing process, then SetFlags
	//is only called once. Thus SetFlags should be idempotent and set static
	//values on f.
	GlobalFlags cli.FlagSetter

	//DisallowGlobalFlagsWithSubCommand denotes whether or not global flags may
	//be present once the sub-command has been defined in the arguments.
	//
	//For example, a false value would allow the following arguments:
	//	["-global1" "1" "sub-command" "-global2" "2" "-subcommand1" "sc1"]
	//while true would not. A true value would require the above ["-global2" "2"]
	//to come before "sub-command" in the argument slice.
	DisallowGlobalFlagsWithSubCommand bool

	names   map[string]SubCommand
	aliases map[string]SubCommand
}

//RegisterHelp registers a help SubCommand that prints out help information about
//a required sub-command parameter.
//The SubCommand's name, synopsis, description, and aliases are provided as parameters.
//If synopsis or description are the empty string, then defaults are used.
func (sc *SubCommander) RegisterHelp(name, synopsis, description string, aliases ...string) {
	help := &helpSubCommand{
		sc: sc,
	}

	if synopsis == "" {
		synopsis = fmt.Sprintf("Prints help information for a %v", SubCommandName)
	}
	if description == "" {
		description = fmt.Sprintf(
			"%v. This includes usage information about the %v's %v and %v",
			synopsis,
			SubCommandName,
			ParametersName,
			SubCommandOptionsName,
		)
	}

	sc.Register(
		&SubCommandStruct{
			NameValue:        name,
			AliasesValue:     aliases,
			SynopsisValue:    synopsis,
			DescriptionValue: description,
			ParameterSetter:  help,
			ExecuteValue:     help.Execute,
		},
	)
}

//RegisterList registers a list SubCommand that prints out all available
//sub-commands when invoked.
//The SubCommand's name, synopsis, description, and aliases are provided as parameters.
//If synopsis or description or the empty string, then defaults are used.
func (sc *SubCommander) RegisterList(name, synopsis, description string, aliases ...string) {
	list := &listSubCommand{
		sc: sc,
	}

	if synopsis == "" {
		synopsis = fmt.Sprintf("Prints available %vs", SubCommandName)
	}
	if description == "" {
		description = synopsis
	}

	sc.Register(
		&SubCommandStruct{
			NameValue:        name,
			AliasesValue:     aliases,
			SynopsisValue:    synopsis,
			DescriptionValue: description,
			ParameterSetter:  list,
			ExecuteValue:     list.Execute,
		},
	)
}

//Register registers subCommand to be possibly executed later via its Name() or
//Aliases().
//This will overwrite any previously registered SubCommands with the same Name()s
//or Aliases().
func (sc *SubCommander) Register(subCommand SubCommand) {
	if sc.names == nil {
		sc.names = map[string]SubCommand{}
	}
	if sc.aliases == nil {
		sc.aliases = map[string]SubCommand{}
	}

	sc.names[subCommand.Name()] = subCommand
	for _, alias := range subCommand.Aliases() {
		sc.aliases[alias] = subCommand
	}
}

//Execute is syntactic sugar for ExecuteContext() with context.Background().
func (sc *SubCommander) Execute(args []string) error {
	return sc.ExecuteContext(context.Background(), args)
}

//ExecuteContext is syntactic sugar for ExecuteContextOut() with os.Stdout and os.Stderr.
func (sc *SubCommander) ExecuteContext(ctx context.Context, args []string) error {
	return sc.ExecuteContextOut(ctx, args, os.Stdout, os.Stderr)
}

//ExecuteContextOut attempts to find and execute a registered SubCommand.
//ctx will be passed along unaltered to the SubCommand's Execute() method.
//args are the command line arguments to parse and use for SubCommand execution.
//They should include command line arguments excluding the program name - usually os.Args[1:].
//out and outErr are the io.Writers to use for standard out and standard error
//for SubCommand execution and help and error output.
//
//TODO
func (sc *SubCommander) ExecuteContextOut(ctx context.Context, args []string, out, outErr io.Writer) (err error) {
	var subCommand SubCommand = nil
	subCommand, err = sc.executeContextOut(ctx, args, out, outErr)
	if err == nil {
		return
	}

	if pgfe, ok := err.(*ParsingGlobalArgsError); ok {
		if pgfe.Err == flag.ErrHelp {
			sc.printCommandError(outErr, nil, true)
		} else {
			sc.printCommandError(outErr, pgfe, true)
		}
		return
	}

	if err == ErrUnsuppliedSubCommand {
		sc.printCommandError(outErr, err, false)
		return
	}

	if _, ok := err.(UnknownSubCommandError); ok {
		sc.printCommandError(outErr, err, false)
		return
	}

	if psce, ok := err.(*ParsingSubCommandError); ok {
		if psce.Err == flag.ErrHelp {
			printSubCommandHeaderDescription(outErr, subCommand)
			fmt.Fprintf(outErr, "%s", "\n\n")
			sc.printSubCommandError(outErr, nil, subCommand)
		} else {
			sc.printSubCommandError(outErr, err, subCommand)
		}
		return
	}

	if _, ok := err.(*ExecutingSubCommandError); ok {
		return
	}

	return
}

func (sc *SubCommander) executeContextOut(ctx context.Context, args []string, out, outErr io.Writer) (SubCommand, error) {
	f := cli.NewFlagSet("", sc.GlobalFlags)
	if err := f.Parse(args); err != nil {
		return nil, &ParsingGlobalArgsError{err}
	}

	args = f.Args()
	if len(args) == 0 {
		return nil, ErrUnsuppliedSubCommand
	}
	name := args[0]
	args = args[1:]

	subCommand := sc.getSubCommand(name)
	if subCommand == nil {
		return nil, UnknownSubCommandError(name)
	}

	return subCommand, sc.executeSubCommand(ctx, f, subCommand, args, out, outErr)
}

func (sc *SubCommander) getSubCommand(name string) SubCommand {
	if subCommand, ok := sc.names[name]; ok {
		return subCommand
	}
	if subCommand, ok := sc.aliases[name]; ok {
		return subCommand
	}
	return nil
}

func (sc *SubCommander) executeSubCommand(
	ctx context.Context,
	gfs *flag.FlagSet,
	subCommand SubCommand,
	args []string,
	out, outErr io.Writer,
) (err error) {
	defer func() {
		if err != nil {
			return
		}
		if r := recover(); r != nil {
			err = &ParsingSubCommandError{fmt.Errorf("%v", r)}
		}
	}()

	err = sc.parseSubCommandArgs(subCommand, args)
	if err != nil {
		err = &ParsingSubCommandError{err}
		return
	}

	err = subCommand.Execute(ctx, out, outErr)
	if err != nil {
		err = &ExecutingSubCommandError{err}
	}

	return
}

func (sc *SubCommander) parseSubCommandArgs(subCommand SubCommand, args []string) error {
	f := cli.NewFlagSet(subCommand.Name(), subCommand)
	if !sc.DisallowGlobalFlagsWithSubCommand && sc.GlobalFlags != nil {
		sc.GlobalFlags.SetFlags(f)
	}

	params, err := cli.ParseArgumentsInterspersed(f, args)
	if err != nil {
		return err
	}

	return subCommand.SetParameters(params)
}

func (sc *SubCommander) printCommandError(out io.Writer, err error, globals bool) {
	if err != nil {
		fmt.Fprintf(out, "%v\n\n", err)
	}

	sc.printCommandUsage(out)

	if globals {
		sc.maybePrintGlobalOptionsUsage(out)
	}
	sc.maybePrintAvailableSubCommands(out)
}

func (sc *SubCommander) printCommandUsage(out io.Writer) {
	fmt.Fprintf(out, "%s %s", Usage, sc.CommandName)

	if sc.hasGlobalOptions() {
		fmt.Fprintf(out, " %v", FormatArgument(GlobalOptionsName, true, true))
	}

	fmt.Fprintf(out, " %v", FormatArgument(SubCommandName, false, false))

	sc.maybePrintSubCommandLineUsage(out, nil)

	fmt.Fprintln(out)
}

func (sc *SubCommander) maybePrintGlobalOptionsUsage(out io.Writer) {
	globalFlagsUsage := sc.getGlobalFlagsUsage()
	if len(globalFlagsUsage) > 0 {
		fmt.Fprintf(out, "\n%s\n", globalFlagsUsage)
	}
}

func (sc *SubCommander) maybePrintAvailableSubCommands(out io.Writer) {
	availableSubCommandsUsage := sc.getAvailableSubCommandsUsage()
	if len(availableSubCommandsUsage) > 0 {
		fmt.Fprintf(out, "\n%s\n", availableSubCommandsUsage)
	}
}

func (sc *SubCommander) printSubCommandError(out io.Writer, err error, subCommand SubCommand) {
	if err != nil {
		if err == flag.ErrHelp {
			printSubCommandHeaderDescription(out, subCommand)
		} else {
			fmt.Fprintf(out, "%v", err)
		}
		fmt.Fprintf(out, "%s", "\n\n")
	}

	fmt.Fprintf(out, "%s %s %s", Usage, "...", subCommand.Name())

	sc.maybePrintSubCommandLineUsage(out, subCommand)

	fmt.Fprintln(out)

	hasGlobalOptions, hasSubCommandOptions, hasParameters := sc.getSubCommandUsageStats(subCommand)
	if hasGlobalOptions && !sc.DisallowGlobalFlagsWithSubCommand {
		sc.maybePrintGlobalOptionsUsage(out)
	}
	if hasSubCommandOptions {
		sc.maybePrintSubCommandOptionsUsage(out, subCommand)
	}
	if hasParameters {
		sc.maybePrintParameters(out, subCommand)
	}
}

func (sc *SubCommander) maybePrintSubCommandOptionsUsage(out io.Writer, subCommand SubCommand) {
	f := cli.NewFlagSet(subCommand.Name(), subCommand)
	defaults := cli.GetFlagSetDefaults(f)
	if len(defaults) > 0 {
		fmt.Fprintf(out, "\n%s:\n%s\n", SubCommandOptionsName, defaults)
	}
}

func (sc *SubCommander) maybePrintParameters(out io.Writer, subCommand SubCommand) {
	params, usage := subCommand.ParameterUsage()

	result := cli.FormatParameters(params, FormatParameter)
	if len(usage) > 0 {
		if len(result) > 0 {
			result += "\n"
		}
		result += usage
	}

	if len(result) > 0 {
		fmt.Fprintf(out, "\n%s: %s\n", ParametersName, result)
	}
}

func (sc *SubCommander) maybePrintSubCommandLineUsage(out io.Writer, subCommand SubCommand) {
	subCommandLineUsage := sc.getSubCommandLineUsage(subCommand)
	if len(subCommandLineUsage) > 0 {
		fmt.Fprintf(out, " %s", subCommandLineUsage)
	}
}

func (sc *SubCommander) getSubCommandLineUsage(subCommand SubCommand) string {
	hasGlobalOptions, hasSubCommandOptions, hasParameters := sc.getSubCommandUsageStats(subCommand)

	args := []string{}
	if hasGlobalOptions && !sc.DisallowGlobalFlagsWithSubCommand {
		args = append(args, GlobalOptionsName)
	}
	if hasSubCommandOptions {
		args = append(args, SubCommandOptionsName)
	}
	if hasParameters {
		args = append(args, ParametersName)
	}

	if len(args) == 0 {
		return ""
	}
	joined := strings.Join(args, ArgumentSeparator)
	if len(args) == 1 {
		return FormatArgument(joined, true, true)
	}
	return FormatArgument(FormatArgument(joined, true, false), true, true)
}

func (sc *SubCommander) getSubCommandUsageStats(subCommand SubCommand) (hasGlobalOptions, hasSubCommandOptions, hasParameters bool) {
	hasGlobalOptions = sc.hasGlobalOptions()
	hasSubCommandOptions = true
	hasParameters = true

	if subCommand != nil {
		params, _ := subCommand.ParameterUsage()
		hasParameters = len(params) > 0

		hasSubCommandOptions = subCommandFlagCount(subCommand) > 0
	}

	return
}

func (sc *SubCommander) hasGlobalOptions() bool {
	return cli.CountFlags(sc.globalFlagSet()) > 0
}

func (sc *SubCommander) globalFlagSet() *flag.FlagSet {
	return cli.NewFlagSet("", sc.GlobalFlags)
}

func (sc *SubCommander) getGlobalFlagsUsage() string {
	defaults := cli.GetFlagSetDefaults(sc.globalFlagSet())
	if len(defaults) == 0 {
		return ""
	}

	return fmt.Sprintf("%s:\n%s", GlobalOptionsName, defaults)
}

func (sc *SubCommander) getAvailableSubCommandsUsage() string {
	if len(sc.names) == 0 {
		return ""
	}

	out := bytes.NewBuffer([]byte{})
	fmt.Fprintf(out, "%s:", SubCommandsName)

	names := sc.sortedSubCommandNames()

	allNameAliases := make([]string, 0, len(names))
	for _, name := range names {
		subCommand := sc.names[name]
		allNameAliases = append(
			allNameAliases,
			getSortedJoinedSubCommandNameAliases(subCommand),
		)
	}

	pad := int(math.Max(16, float64(maxLen(allNameAliases)+4)))
	for i, name := range names {
		sc := sc.names[name]
		nameAliases := allNameAliases[i]
		fmt.Fprintf(out, "\n  %s%s%s", nameAliases, padRight(pad, nameAliases), sc.Synopsis())
	}

	return out.String()
}

func printSubCommandHeaderDescription(out io.Writer, subCommand SubCommand) {
	fmt.Fprintf(
		out,
		"%s",
		getSortedJoinedSubCommandNameAliases(subCommand),
	)
	if description := subCommand.Description(); len(description) > 0 {
		fmt.Fprintf(out, " - %s", description)
	}
}

func getSortedJoinedSubCommandNameAliases(subCommand SubCommand) string {
	return cli.GetJoinedNameSortedAliases(subCommand.Name(), subCommand.Aliases())
}

func (sc *SubCommander) sortedSubCommandNames() []string {
	names := make([]string, 0, len(sc.names))
	for name := range sc.names {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func maxLen(values []string) int {
	max := 0
	for _, value := range values {
		if l := len(value); l > max {
			max = l
		}
	}
	return max
}

func padRight(count int, value string) string {
	count = count - len(value)
	result := make([]byte, count)
	for i := range result {
		result[i] = ' '
	}
	return string(result)
}

type helpSubCommand struct {
	sc *SubCommander

	helpSubCommandName string
}

func (h *helpSubCommand) ParameterUsage() ([]*cli.Parameter, string) {
	params := []*cli.Parameter{
		{Name: SubCommandName, Optional: false, Many: false},
	}
	usage := fmt.Sprintf("%v is the %v to provide help for", FormatParameter(params[0]), SubCommandName)

	return params, usage
}

func (h *helpSubCommand) SetParameters(params []string) error {
	if len(params) > 1 {
		return cli.ErrTooManyParameters
	}
	if len(params) == 0 {
		return &cli.RequiredParameterNotSetError{
			Name: SubCommandName,
			Many: false,
			Formatted: FormatParameter(
				&cli.Parameter{
					Name: SubCommandName,
					Many: false,
				},
			),
		}
	}

	h.helpSubCommandName = params[0]
	return nil
}

func (h *helpSubCommand) Execute(_ context.Context, out, outErr io.Writer) error {
	subCommand := h.sc.getSubCommand(h.helpSubCommandName)
	if subCommand == nil {
		err := UnknownSubCommandError(h.helpSubCommandName)
		h.sc.printCommandError(outErr, err, false)
		return err
	}

	h.sc.printSubCommandError(out, flag.ErrHelp, subCommand)

	return nil
}

type listSubCommand struct {
	sc *SubCommander
}

func (l *listSubCommand) ParameterUsage() ([]*cli.Parameter, string) {
	return nil, ""
}

func (l *listSubCommand) SetParameters(params []string) error {
	if len(params) > 0 {
		return cli.ErrTooManyParameters
	}
	return nil
}

func (l *listSubCommand) Execute(_ context.Context, out, _ io.Writer) error {
	fmt.Fprintf(out, "%s\n", l.sc.getAvailableSubCommandsUsage())
	return nil
}