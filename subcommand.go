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
	//just as if the SubCommand's name used from the command line.
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

	ParameterUsage(format func(p *Parameter) string) string

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
	GlobalFlags FlagSetter

	DisallowGlobalFlagsWithSubCommand bool

	ParameterFlagMode

	names   map[string]SubCommand
	aliases map[string]SubCommand
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

func (sc *SubCommander) ExecuteContextOut(ctx context.Context, args []string, out, outErr io.Writer) (err error) {
	var subCommand SubCommand = nil

	subCommand, err = sc.executeContextOut(ctx, args, out, outErr)
	if err == nil {
		return
	}

	if epgf, ok := err.(ErrParsingGlobalArgs); ok {
		if epgf == flag.ErrHelp {
			fmt.Println("global flag error help")
		} else {
			fmt.Println("global flag other parsing failed")
		}
		return
	}

	if err == ErrUnsuppliedSubCommand {
		fmt.Println("error unsupplied subcommand")
		return
	}

	if _, ok := err.(ErrUnknownSubCommand); ok {
		fmt.Println("error unknown subcommand")
	}

	if _, ok := err.(ErrParsingSubCommand); ok {
		fmt.Println("error parsing subcommand")
	}

	if _, ok := err.(ErrExecutingSubCommand); ok {
		return
	}

	fmt.Println(subCommand.Name())

	return
}

func (sc *SubCommander) executeContextOut(ctx context.Context, args []string, out, outErr io.Writer) (SubCommand, error) {
	f := newFlagSet("")

	if sc.GlobalFlags != nil {
		sc.GlobalFlags.SetFlags(f)
	}
	if err := f.Parse(args); err != nil {
		return nil, ErrParsingGlobalArgs(err)
	}

	args = f.Args()
	if len(args) == 0 {
		return nil, ErrUnsuppliedSubCommand
	}
	name := args[0]
	args = args[1:]

	subCommand := sc.getSubCommand(name)
	if subCommand == nil {
		return nil, ErrUnknownSubCommand(name)
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
			err = ErrParsingSubCommand(fmt.Errorf("%v", r))
		}
	}()

	err = sc.parseSubCommandArgs(subCommand, args)
	if err != nil {
		err = ErrParsingSubCommand(err)
		return
	}

	err = subCommand.Execute(ctx, out, outErr)
	if err != nil {
		err = ErrExecutingSubCommand(err)
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
		return ErrFlagsAfterParameters(err.Error())
	}

	if len(args) > 0 {
		return ErrFlagsAfterParameters(strings.Join(args, ", "))
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
