package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

//SubCommand defines the options for a sub-command to be executed by a SubCommander.
type SubCommand interface {
	//Name returns the name of the SubCommand.
	//Names must not overlap with other SubCommand Name()s or Aliases() in the
	//same SubCommander.
	Name() string

	//Aliases returns the aliases of the SubCommand.
	//These are listed in command help output and will cause a SubCommand to execute
	//just as if the SubCommand's name were used from the command line.
	//Aliases must not overlap with other SubCommand Name()s or Aliases() in the
	//same SubCommander.
	Aliases() []string

	//Synopsis returns a synopsis of the SubCommand.
	//This should be less than one line and is used for command help output.
	Synopsis() string

	//Description returns a longer description of the SubCommand.
	//This can be however long you like and is used for help specific to this
	//SubCommand.
	Description() string

	//FlagSetter for SubCommand flags.
	FlagSetter

	//ParameterUsage returns the Parameters used by the SubCommand and a possible
	//usage string to describe params in more detail.
	//These values are used in help and error output.
	ParameterUsage() (params []*Parameter, usage string)

	//SetParameters allows SubCommands to receive parameter arguments during argument
	//parsing.
	SetParameters(values []string) error

	//Execute is where the SubCommand should do its work.
	//A non-nil return value indicates the execution failed and that error will
	//be process by a SubCommander.
	Execute(ctx context.Context, out, outErr io.Writer) error
}

//SubCommander provides registering multiple SubCommands and executing them from
//command line arguments.
//
//Note that SubCommander is NOT safe for use with multiple goroutines.
type SubCommander struct {
	//CommandName is used in error and help output. It should be the name of the
	//program that was invoked.
	//It is set to os.Args[0] by NewSubCommander().
	CommandName string

	//GlobalFlags is a FlagSetter that is used for setting global flags for subcommands.
	GlobalFlags FlagSetter

	//DisallowGlobalFlagsWithSubCommand denotes whether or not global flags may
	//be present once the sub-command has been defined in the arguments.
	//
	//For example, a false value would allow the following arguments:
	//	["-global1" "1" "sub-command" "-global2" "2" "-subcommand1" "sc1"]
	//while true would not. A true value would require the above ["-global2" "2"]
	//to come before "sub-command" in the argument slice.
	DisallowGlobalFlagsWithSubCommand bool

	//ParameterFlagMode is the mode used when parsing sub-command flags (and the
	//possible inclusion of global flags depending on the value of
	//DisallowGlobalFlagsWithsubCommand) and parameters.
	//
	//See the ParameterFlagMode type for more details. It is set to
	ParameterFlagMode

	names   map[string]SubCommand
	aliases map[string]SubCommand
}

func NewSubCommander() *SubCommander {
	return &SubCommander{
		CommandName: os.Args[0],
	}
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
			NameValue:           name,
			AliasesValue:        aliases,
			SynopsisValue:       synopsis,
			DescriptionValue:    description,
			ParameterUsageValue: help.ParameterUsage,
			SetParametersValue:  help.SetParameters,
			ExecuteValue:        help.Execute,
		},
	)
}

//RegisterListSubCommands registers a list SubCommand that prints out all available
//sub-commands when invoked.
//The SubCommand's name, synopsis, description, and aliases are provided as parameters.
//If synopsis or description or the empty string, then defaults are used.
func (sc *SubCommander) RegisterListSubCommands(name, synopsis, description string, aliases ...string) {
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
			NameValue:           name,
			AliasesValue:        aliases,
			SynopsisValue:       synopsis,
			DescriptionValue:    description,
			ParameterUsageValue: list.ParameterUsage,
			SetParametersValue:  list.SetParameters,
			ExecuteValue:        list.Execute,
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
//They should include all command line arguments including the program name.
//out and outErr are the io.Writers to use for standard out and standard error
//for SubCommand execution and help and error output.
//
//
func (sc *SubCommander) ExecuteContextOut(ctx context.Context, args []string, out, outErr io.Writer) (err error) {
	var subCommand SubCommand = nil

	subCommand, err = sc.executeContextOut(ctx, args, out, outErr)
	if err == nil {
		return
	}

	if epgf, ok := err.(ParsingGlobalArgsError); ok {
		if epgf == flag.ErrHelp {
			fmt.Println("global flag error help")
		} else {
			fmt.Println("global flag other parsing failed")
		}
		return
	}

	if err == ErrUnsuppliedSubCommand {
		sc.printErrUnsuppliedSubCommand(outErr)
		return
	}

	if _, ok := err.(UnknownSubCommandError); ok {
		fmt.Println("error unknown subcommand")
	}

	if _, ok := err.(ParsingSubCommandError); ok {
		fmt.Println("error parsing subcommand")
	}

	if _, ok := err.(ExecutingSubCommandError); ok {
		return
	}

	fmt.Fprintf(ioutil.Discard, subCommand.Name())

	return
}

func (sc *SubCommander) executeContextOut(ctx context.Context, args []string, out, outErr io.Writer) (SubCommand, error) {
	f := newFlagSet("")

	if sc.GlobalFlags != nil {
		sc.GlobalFlags.SetFlags(f)
	}
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
			err = ParsingSubCommandError{fmt.Errorf("%v", r)}
		}
	}()

	err = sc.parseSubCommandArgs(subCommand, args)
	if err != nil {
		err = ParsingSubCommandError{err}
		return
	}

	err = subCommand.Execute(ctx, out, outErr)
	if err != nil {
		err = &ExecutingSubCommandError{err}
	}

	return
}

func (sc *SubCommander) parseSubCommandArgs(subCommand SubCommand, args []string) error {
	f := newFlagSet(subCommand.Name())
	subCommand.SetFlags(f)
	if !sc.DisallowGlobalFlagsWithSubCommand && sc.GlobalFlags != nil {
		sc.GlobalFlags.SetFlags(f)
	}

	switch sc.ParameterFlagMode {
	case ModeFlagsFirst:
		return parseSubCommandFlagsFirst(f, subCommand, args)
	case ModeParametersFirst:
		return parseSubCommandParametersFirst(f, subCommand, args)
	default:
		return parseSubCommandInterspersed(f, subCommand, args)
	}

	return nil
}

func (sc *SubCommander) printErrUnsuppliedSubCommand(out io.Writer) {
	fmt.Fprintf(out, "%v\n\n", ErrUnsuppliedSubCommand)
	fmt.Fprintf(out, UsageFormat, sc.CommandName)
	fmt.Fprintf(out, " %v", sc.CommandName)

	f := sc.globalFlagSet()
	if countFlags(f) > 0 {
		fmt.Fprintf(out, " %v", FormatArgument(GlobalOptionsName, true, true))
	}

	fmt.Fprintf(
		out,
		" %v %v %v\n",
		FormatArgument(SubCommandName, false, false),
		FormatArgument(ParametersName, true, true),
		FormatArgument(SubCommandOptionsName, true, true),
	)

	availableSubCommandUsage := sc.getAvailableSubCommandUsage()
	if len(availableSubCommandUsage) > 0 {
		fmt.Fprintf(out, "\n%s\n", availableSubCommandUsage)
	}
}

func (sc *SubCommander) globalFlagSet() *flag.FlagSet {
	f := newFlagSet("")
	if sc.GlobalFlags != nil {
		sc.GlobalFlags.SetFlags(f)
	}
	return f
}

func (sc *SubCommander) getAvailableSubCommandUsage() string {
	result := fmt.Sprintf(AvailableFormat, SubCommandsName)

	return result
}

func parseSubCommandFlagsFirst(f *flag.FlagSet, subCommand SubCommand, args []string) error {
	if err := f.Parse(args); err != nil {
		return err
	}
	return subCommand.SetParameters(f.Args())
}

func parseSubCommandParametersFirst(f *flag.FlagSet, subCommand SubCommand, args []string) error {
	var err error = nil
	params := []string{}

	for err == nil && f.NFlag() == 0 && len(args) > 0 {
		err = f.Parse(args)
		if err != nil {
			continue
		}
		if didStopAfterDoubleMinus(args, f.Args()) {
			params = append(params, f.Args()...)
			args = args[len(args):]
			continue
		}
		args = f.Args()
		if len(args) > 0 {
			params = append(params, args[0])
			args = args[1:]
		}
	}
	if err != nil {
		if err == flag.ErrHelp {
			return err
		}
		return FlagsAfterParametersError(err.Error())
	}

	if len(args) > 0 {
		return FlagsAfterParametersError(strings.Join(args, ", "))
	}

	return subCommand.SetParameters(params)
}

func parseSubCommandInterspersed(f *flag.FlagSet, subCommand SubCommand, args []string) error {
	var err error = nil
	params := []string{}

	for err == nil && len(args) > 0 {
		err = f.Parse(args)
		if err != nil {
			continue
		}
		if didStopAfterDoubleMinus(args, f.Args()) {
			params = append(params, f.Args()...)
			args = args[len(args):]
			continue
		}
		args = f.Args()
		if len(args) > 0 {
			params = append(params, args[0])
			args = args[1:]
		}
	}
	if err != nil {
		return err
	}

	return subCommand.SetParameters(params)
}

func didStopAfterDoubleMinus(args, remaining []string) bool {
	return len(args) > len(remaining) && args[len(args)-len(remaining)-1] == DoubleMinus
}

func newFlagSet(name string) *flag.FlagSet {
	f := flag.NewFlagSet(name, flag.ContinueOnError)
	f.Usage = func() {}
	f.SetOutput(ioutil.Discard)
	return f
}

func countFlags(f *flag.FlagSet) int {
	count := 0
	f.VisitAll(func(_ *flag.Flag) {
		count++
	})
	return count
}

type helpSubCommand struct {
	sc *SubCommander

	helpCommandName string
}

func (h *helpSubCommand) ParameterUsage() ([]*Parameter, string) {
	params := []*Parameter{
		{Name: SubCommandName, Optional: false, Many: false},
	}
	usage := fmt.Sprintf("%v is the %v to provide help for", FormatParameterName(params[0].Name), SubCommandName)

	return params, usage
}

func (h *helpSubCommand) SetParameters(params []string) error {
	if len(params) > 1 {
		return ErrInvalidParameters
	}
	if len(params) == 0 {
		return &RequiredParameterNotSetError{
			Name: SubCommandName,
			Many: false,
		}
	}

	h.helpCommandName = params[0]
	return nil
}

func (h *helpSubCommand) Execute(_ context.Context, out, _ io.Writer) error {
	_, err := fmt.Fprintln(out, "help execute unimplemented")
	return err
}

type listSubCommand struct {
	sc *SubCommander
}

func (l *listSubCommand) ParameterUsage() ([]*Parameter, string) {
	return nil, NoParametersUsage
}

func (l *listSubCommand) SetParameters(params []string) error {
	if len(params) != 0 {
		return fmt.Errorf(NoParametersUsage)
	}
	return nil
}

func (l *listSubCommand) Execute(_ context.Context, out, _ io.Writer) error {
	_, err := fmt.Fprintln(out, "list execute unimplemented")
	return err
}
