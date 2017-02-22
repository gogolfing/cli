package cli

import "flag"

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
