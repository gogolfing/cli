//Package command provides a Command interface and a Commander type that is able
//to execute a command line application that does "one" thing, i.e. without
//subcommands.
//
//The Command type provides methods needed for parsing and setting command line
//arguments and executing code once parsing is complete.
//
//The Commander type is a simple wrapper that handles executing a Command and providing
//help and error output in the case of errors no related to executing a command.
//
//The help and error output follow the general form loosely based on Go templates.
//	{{.ErrorIfAParsingErrorNotAnExecutionError}}
//
//	{{.CommandNameAndDescriptionIfErrHelp}}
//
//	usage: {{.AvailableCommandLineArguments}}
//
//	{{.AvailableFlagOptionUsageIfThereAreOptions}}
//
//	{{.AvailableParameterUsageIfThereAreParameters}}
package command