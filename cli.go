package cli

import "flag"

//DoubleMinus is the argument to determine if the flag package has stopped parsing
//after seeing this argument.
const DoubleMinus = "--"

//Output values that affect error and help output.
//
//These variables may be changed to affect the output of this package.
var (
	//UsageFormat is used to format the usage line of help and error output.
	//It should have exactly one format argument that is the command or sub-command.
	UsageFormat = "Usage of %v:"

	//AvailableFormat is used the format available flag and parameter usage.
	//It should have exactly one format argument that is the type being described.
	AvailableFormat = "Available %v:"

	ParameterName         = "parameter"
	ParametersName        = "parameters"
	GlobalOptionsName     = "global_options"
	SubCommandName        = "sub_command"
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
