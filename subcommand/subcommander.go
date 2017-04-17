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
	//Usually os.Args[0].
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
	if synopsis == "" {
		synopsis = fmt.Sprintf("Prints help information for a %v", SubCommandName)
	}
	if description == "" {
		description = fmt.Sprintf(
			"%v. This includes usage information about the %v's %v and %v.",
			synopsis,
			SubCommandName,
			ParametersName,
			SubCommandOptionsName,
		)
	}

	sc.Register(
		&helpSubCommand{
			sc: sc,
			SubCommandStruct: &SubCommandStruct{
				NameValue:        name,
				AliasesValue:     aliases,
				SynopsisValue:    synopsis,
				DescriptionValue: description,
			},
		},
	)
}

//RegisterList registers a list SubCommand that prints out all available
//sub-commands when invoked.
//The SubCommand's name, synopsis, description, and aliases are provided as parameters.
//If synopsis or description or the empty string, then defaults are used.
func (sc *SubCommander) RegisterList(name, synopsis, description string, aliases ...string) {
	if synopsis == "" {
		synopsis = fmt.Sprintf("Prints available %vs", SubCommandName)
	}
	if description == "" {
		description = synopsis + "."
	}

	sc.Register(
		&listSubCommand{
			sc: sc,
			SubCommandStruct: &SubCommandStruct{
				NameValue:        name,
				AliasesValue:     aliases,
				SynopsisValue:    synopsis,
				DescriptionValue: description,
			},
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

//Execute is syntactic sugar for sc.ExecuteContext() with context.Background(), args,
//os.Stdin, os.Stdout, and os.Stderr.
func (sc *SubCommander) Execute(args []string) error {
	return sc.ExecuteContext(context.Background(), args, os.Stdin, os.Stdout, os.Stderr)
}

//ExecuteContext executes a SubCommand registered with sc with the provided parameters.
//
//Ctx is the Context passed unaltered to SubCommand.Execute.
//
//Args should be the program arguments excluding the program name - usually os.Args[1:].
//
//The parameters in, out, and outErr are passed unaltered to SubCommand.Execute
//and should represent the standard input, output, and error files for the executing
//SubCommand.
//
//Err will be non-nil if parsing args failed - with type *ParsingGlobalArgsError,
//*ParsingSubCommandError.
//It will be ErrUnsuppliedSubCommand if the subcommand is not supplied in command
//line arguments.
//It will be of type UnknownSubCommandError if the subcommand name arguments supplied
//was not found in the registered SubCommands' names or aliases.
//It will be of type *ExecutingSubCommandError if the SubCommand.Execute
//method returns an error.
//
//If the returned error is of type *ParsingGlobalArgsError, *ParsingSubCommandError,
//ErrUnsuppliedSubCommand, or UnknownSubCommandError then error and help output
//will be written to outErr.
//See the package documentation for more details on error and help output.
//If this is the error, then execution stops and SubCommand.Execute is never called.
//
//If the error is an *ExecutingSubCommandError then nothing is output by sc.
func (sc *SubCommander) ExecuteContext(ctx context.Context, args []string, in io.Reader, out, outErr io.Writer) (err error) {
	var subCommand SubCommand
	subCommand, err = sc.executeContext(ctx, args, in, out, outErr)
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
			sc.printSubCommandError(outErr, nil, true, subCommand)
		} else {
			sc.printSubCommandError(outErr, err, true, subCommand)
		}
		return
	}

	if _, ok := err.(*ExecutingSubCommandError); ok {
		return
	}

	return
}

func (sc *SubCommander) executeContext(ctx context.Context, args []string, in io.Reader, out, outErr io.Writer) (SubCommand, error) {
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

	return subCommand, sc.executeSubCommand(ctx, f, subCommand, args, in, out, outErr)
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
	f *flag.FlagSet,
	subCommand SubCommand,
	args []string,
	in io.Reader,
	out, outErr io.Writer,
) (err error) {
	err = sc.parseSubCommandArgs(subCommand, f, args)
	if err != nil {
		err = &ParsingSubCommandError{err}
		return
	}

	err = subCommand.Execute(ctx, in, out, outErr)
	if err != nil {
		err = &ExecutingSubCommandError{err}
	}

	return
}

func (sc *SubCommander) parseSubCommandArgs(subCommand SubCommand, gf *flag.FlagSet, args []string) error {
	f := gf
	if sc.DisallowGlobalFlagsWithSubCommand {
		f = cli.NewFlagSet(subCommand.Name(), nil)
	}
	if fs, ok := subCommand.(cli.FlagSetter); ok {
		fs.SetFlags(f)
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

	sc.maybePrintSubCommandLineUsage(out, nil, true)

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

func (sc *SubCommander) printSubCommandError(out io.Writer, err error, globals bool, subCommand SubCommand) {
	if err != nil {
		if err == flag.ErrHelp {
			printSubCommandHeaderDescription(out, subCommand)
		} else {
			fmt.Fprintf(out, "%v", err)
		}
		fmt.Fprintf(out, "%s", "\n\n")
	}

	fmt.Fprintf(out, "%s %s %s", Usage, "...", subCommand.Name())

	sc.maybePrintSubCommandLineUsage(out, subCommand, globals)

	fmt.Fprintln(out)

	hasGlobalOptions, hasSubCommandOptions, _ := sc.getSubCommandUsageStats(subCommand)
	if globals && hasGlobalOptions && !sc.DisallowGlobalFlagsWithSubCommand {
		sc.maybePrintGlobalOptionsUsage(out)
	}
	if hasSubCommandOptions {
		sc.maybePrintSubCommandOptionsUsage(out, subCommand)
	}
	sc.maybePrintParameters(out, subCommand)
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

	didPrint := false
	if formatted := cli.FormatParameters(params, FormatParameter); len(formatted) > 0 {
		fmt.Fprintf(out, "\n%s: %s", ParametersName, formatted)
		didPrint = true
	}
	if len(usage) > 0 {
		fmt.Fprintf(out, "\n%s", usage)
		didPrint = true
	}
	if didPrint {
		fmt.Fprintln(out)
	}
}

func (sc *SubCommander) maybePrintSubCommandLineUsage(out io.Writer, subCommand SubCommand, globals bool) {
	subCommandLineUsage := sc.getSubCommandLineUsage(subCommand, globals)
	if len(subCommandLineUsage) > 0 {
		fmt.Fprintf(out, " %s", subCommandLineUsage)
	}
}

func (sc *SubCommander) getSubCommandLineUsage(subCommand SubCommand, globals bool) string {
	hasGlobalOptions, hasSubCommandOptions, hasParameters := sc.getSubCommandUsageStats(subCommand)

	args := []string{}
	if globals && hasGlobalOptions && !sc.DisallowGlobalFlagsWithSubCommand {
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

		hasSubCommandOptions = cli.CountFlags(cli.NewFlagSet(subCommand.Name(), subCommand)) > 0
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

	*SubCommandStruct
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

func (h *helpSubCommand) Execute(_ context.Context, _ io.Reader, out, outErr io.Writer) error {
	subCommand := h.sc.getSubCommand(h.helpSubCommandName)
	if subCommand == nil {
		err := UnknownSubCommandError(h.helpSubCommandName)
		h.sc.printCommandError(outErr, err, false)
		return err
	}

	_, helpOk := subCommand.(*helpSubCommand)
	_, listOk := subCommand.(*listSubCommand)

	h.sc.printSubCommandError(out, flag.ErrHelp, !helpOk && !listOk, subCommand)

	return nil
}

type listSubCommand struct {
	sc *SubCommander

	*SubCommandStruct
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

func (l *listSubCommand) Execute(_ context.Context, _ io.Reader, out, _ io.Writer) error {
	fmt.Fprintf(out, "%s\n", l.sc.getAvailableSubCommandsUsage())
	return nil
}
