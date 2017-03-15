package cli

import (
	"context"
	"io"
)

//SubCommand defines the options for a sub-command to be executed by a SubCommander.
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
	FlagSetter

	//ParameterUsage returns the Parameters used by the SubCommand and a possible
	//usage string to describe params in more detail.
	//These values are used in help and error output.
	ParameterUsage() (params []*Parameter, usage string)

	//SetParameters allows SubCommands to receive parameter arguments during argument
	//parsing.
	//Implementations should not retain references to values.
	SetParameters(values []string) error

	//Execute is where the SubCommand should do its work.
	//A non-nil return value indicates the execution failed and that error will
	//be process by a SubCommander.
	Execute(ctx context.Context, out, outErr io.Writer) error
}
