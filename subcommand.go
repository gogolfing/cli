package cli

import (
	"context"
	"io"
	"os"
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

	//Parameters returns the parameters of the SubCommand.
	//The order returned is important and will be enforced in parsing.
	Parameters() []*Parameter

	//ParameterUsage returns the usage text for the SubCommand.
	//This should not include the SubCommand's name or aliases.
	//An empty return value means that help output will be formatted from the
	//result of Parameters(). A non-empty value will be used directly for help output.
	ParameterUsage(params []*Parameter, format func(p *Parameter) string) string

	//SetParameter allows the SubCommand to receive parameter values when parsing.
	SetParameter(index int, value string) error

	//ParameterFlagMode returns the ParameterFlagMode the SubCommand will use
	//when parsing.
	ParameterFlagMode() ParameterFlagMode

	//ValidateParameters should ensure that the set parameters are valid.
	//This will be called after successful sub-command argument parsing to
	//allow SubCommands to ensure parameters were set properly.
	//I.e. to ensure required parameters are actually set.
	ValidateParameters() error

	//FlagSetter for SubCommand flags.
	FlagSetter

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

func (sc *SubCommander) Execute(args []string) error {
	return sc.ExecuteContext(context.Background(), args)
}

func (sc *SubCommander) ExecuteContext(ctx context.Context, args []string) error {
	return sc.ExecuteContextOut(ctx, args, os.Stdout, os.Stderr)
}

func (sc *SubCommander) ExecuteContextOut(ctx context.Context, args []string, out, outErr io.Writer) error {
	err := sc.executeContextOut(ctx, args, out, outErr)

	//do processing on err here.

	return err
}

func (sc *SubCommander) executeContextOut(ctx context.Context, args []string, out, outErr io.Writer) error {
	if len(args) < 1 {
		return ErrUnsuppliedSubCommand
	}

	//need to actually parse global flags and get subcommand from Args().

	name := args[0]
	args = args[1:]

	subCommand := sc.getSubCommand(name)
	if subCommand == nil {
		return ErrUnknownSubCommand(name)
	}

	return sc.executeSubCommand(ctx, subCommand, out, outErr)
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

func (sc *SubCommander) executeSubCommand(ctx context.Context, subCommand SubCommand, out, outErr io.Writer) error {
	return nil
}
