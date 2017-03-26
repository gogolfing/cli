package subcommand

import (
	"context"
	"io"

	"github.com/gogolfing/cli"
)

//Output values that affect error and help output.
const (
	Usage = cli.Usage

	GlobalOptionsName = "global_options"

	SubCommandName        = "sub_command"
	SubCommandsName       = "sub_commands"
	SubCommandOptionsName = "sub_command_options"

	ParametersName = cli.ParametersName

	ArgumentSeparator = cli.ArgumentSeparator
)

//ForamtArgument formats an argument's name given whether or not it is optional
//or multiple values are allowed.
//
//This value may be changed to affect the output of this package.
var FormatArgument = cli.FormatArgument

//FormatParameter returns a string representation of p appropriate for help and
//error output.
//
//This value may be changed to affect the output of this package.
var FormatParameter = cli.FormatParameter

//FormatParameterName returns a string representation of a Parameter name appropriate
//for help and error output.
//
//This value my be changed to affect the output of this package.
var FormatParameterName = cli.FormatParameterName

//SubCommand defines the options for a sub-command to be executed by a SubCommander.
//
//The methods of this interface (except SetParameters and Execute) may be called
//multiple times to obtain help and error output when needed. These methods should
//be idempotent.
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
	//
	//This value may be called multiple times with different values for f if argument
	//parsing fails. This is done to obtain possible help and error output.
	//If there are no errors during the argument parsing process, then SetFlags
	//is only called once. Thus SetFlags should be idempotent and set static
	//values on f.
	cli.FlagSetter

	//ParameterSetter for SubCommand parameters.
	cli.ParameterSetter

	//Execute is where the SubCommand should do its work.
	//A non-nil return value indicates the execution failed and that error will
	//be processed by a SubCommander.
	Execute(ctx context.Context, out, outErr io.Writer) error
}

func subCommandFlagCount(subCommand SubCommand) int {
	return cli.CountFlags(cli.NewFlagSet(subCommand.Name(), subCommand))
}
