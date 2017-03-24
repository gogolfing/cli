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

	ParameterName  = "parameter"
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

type Command interface {
	Description() string

	cli.FlagSetter

	cli.ParameterSetter

	Execute(ctx context.Context, out, outErr io.Writer) error
}
