//Package subcommand provides a SubCommand interface and a SubCommander type that
//is able to execute a command line application that has multiple subcommands.
//
//The SubCommand type provides methods needed ofr parsing and setting command line
//arguments and executing code once parsing is complete.
//
//The SubCommander type is a wrapper that handles executing a SubCommand and providing
//help and error output in the case of errors not related to executing a SubCommand.
//
//The help and error output follow the general form loosely based on Go templates:
//	{{.ErrorIfAParsingErrorNotAnExecutionError}}
//
//	{{if .KnowTheSubCommand}}
//		{{.SubCommandNameAndDescriptionIfErrHelp}}
//
//		usage: {{.SubCommander.CommandName}} {{.AvailableCommandLineArguments}}
//
//		{{.AvailableFlagOptionUsageIfThereAreOptions}}
//
//		{{.AvailableParameterUsageIfThereAreParameters}}
//	{{else}}
//		{{.GlobalOptionUsageIfThereAreGlobalOptions}}
//
//		{{.AllSubCommandUsage}}
//	{{end}}
//
//Please see the examples for actual output.
package subcommand
