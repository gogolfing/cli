package cli

import "fmt"

//ErrExitStatus allows wrapping together an exit status with an error providing
//that exist code.
type ErrExitStatus struct {
	//Code is the desired exit status.
	Code int

	//error is the wrapped error.
	error
}

//ErrInvalidParameters is a generic error for invalid parameters being set.
var ErrInvalidParameters = fmt.Errorf("invalid %v", ParametersName)

//ErrRequiredParameterNotSet is an error that denotes a parameter was not supplied
//in the arguments but is required to be present.
type ErrRequiredParameterNotSet struct {
	Name string
	Many bool
}

//Error provides the error implementation.
func (e *ErrRequiredParameterNotSet) Error() string {
	return fmt.Sprintf(
		"required %v %v not set",
		ParameterName,
		FormatParameter(&Parameter{Name: e.Name, Many: e.Many}),
	)
}

//ErrUnsuppliedSubCommand is a value error denoting a sub-command was not supplied
//during argument parsing.
var ErrUnsuppliedSubCommand = fmt.Errorf("%v not supplied", SubCommandName)

//ErrUnknownSubCommand is an error denoting the provided sub-command is not registered.
type ErrUnknownSubCommand string

//Error provides the error implementation.
func (e ErrUnknownSubCommand) Error() string {
	return fmt.Sprintf("unknown %v %q", SubCommandName, string(e))
}

//ErrParsingGlobalArgs is an error wrapper denoting global argument parsing failed.
type ErrParsingGlobalArgs struct {
	error
}

//ErrParsingSubCommand is an error wrapper denoting sub-command argument parsing
//failed.
type ErrParsingSubCommand struct {
	error
}

//ErrFlagsAfterParameters is an error denoting flags were parsed after the parameter
//arguments. Note that this error will be used only with specific ParameterFlagModes.
type ErrFlagsAfterParameters string

//Error provides the error implementation.
func (e ErrFlagsAfterParameters) Error() string {
	return fmt.Sprintf("flags present after %v: %v", ParametersName, string(e))
}

//ErrExecutingSubCommand is an error wrapper denoting that executing a sub-command
//failed.
type ErrExecutingSubCommand struct {
	error
}

//IsExecutionError returns whether or not err is an ErrExecutingSubCommand error.
func IsExecutionError(err error) bool {
	switch err.(type) {
	case ErrExecutingSubCommand, *ErrExecutingSubCommand:
		return true
	}
	return false
}
