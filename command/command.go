package command

import (
	"context"
	"io"

	"github.com/gogolfing/cli"
)

//Output values that affect error and help output.
const (
	Usage = "usage:"

	OptionsName = "options"

	ParametersName = "parameters"

	ArgumentSeparator = " | "
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

//Command defines the options for executing a command.
//
//The methods of this interface (except SetParameters and Execute) may be called
//multiple times to obtain help and error output when needed. These methods should
//be idempotent.
type Command interface {
	//Description returns a description of the Command.
	Description() string

	//FlagSetter for Command flags.
	//
	//This value may be called multiple times with different values for f if argument
	//parsing fails. This is done to obtain possible help and error output.
	//If there are no errors during the argument parsing process, then SetFlags
	//is only called once. Thus SetFlags should be idempotent and set static
	//values on f.
	cli.FlagSetter

	//ParameterSetter for Command parameters.
	cli.ParameterSetter

	//Execute is where the Command should do its work.
	//A non-nil return value indicates the execution failed and that error will
	//be processed by a Commander.
	Execute(ctx context.Context, out, outErr io.Writer) error
}
