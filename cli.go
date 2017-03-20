package cli

import "flag"

//DoubleMinus is the argument to determine if the flag package has stopped parsing
//after seeing this argument.
const DoubleMinus = "--"

//Output values that affect error and help output.
const (
	Usage = "usage:"

	GlobalOptionsName = "global_options"

	ParameterName  = "parameter"
	ParametersName = "parameters"

	SubCommandName        = "sub_command"
	SubCommandsName       = "sub_commands"
	SubCommandOptionsName = "sub_command_options"

	ArgumentSeparator = " | "
)

//FlagSetter allows implementations to receive values from flag.FlagSets while
//argument parsing occurs.
//Implementations should not retain references to f.
type FlagSetter interface {
	SetFlags(f *flag.FlagSet)
}

var FormatArgument = func(name string, optional, many bool) string {
	result := name
	if many {
		result += "..."
	}
	if optional {
		result = "[" + result + "]"
	} else {
		result = "<" + result + ">"
	}
	return result
}
