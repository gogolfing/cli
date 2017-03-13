package cli

import "flag"

//DoubleMinus is the argument to determine if the flag package has stopped parsing
//after seeing this argument.
const DoubleMinus = "--"

//Output values that affect error and help output.
const (
	UsageFormat = "Usage of %s:"

	AvailableFormat = "Available %s:"

	GlobalOptionsName = "global_options"

	ParameterName  = "parameter"
	ParametersName = "parameters"

	SubCommandName        = "sub_command"
	SubCommandsName       = "sub_commands"
	SubCommandOptionsName = "sub_command_options"

	NoParametersUsage = "there are no " + ParametersName
)

//ParameterFlagMode determines how arguments are parsed for SubCommands.
type ParameterFlagMode int

const (
	//ModeInterspersed allows command parameters and flags to be mixed with eachother
	//in their ordering.
	ModeInterspersed ParameterFlagMode = iota

	//ModeFlagsFirst requires all flag options to come before parameters.
	ModeFlagsFirst

	//ModeParametersFirst requires all parameters to come before flag options.
	ModeParametersFirst
)

//FlagSetter allows implementations to receive values from flag.FlagSets while
//argument parsing occurs.
type FlagSetter interface {
	SetFlags(f *flag.FlagSet)
}

func FormatArgument(name string, optional, many bool) string {
	result := name
	if many {
		result += " ..."
	}
	if optional {
		result = "[" + result + "]"
	} else {
		result = "<" + result + ">"
	}
	return result
}
