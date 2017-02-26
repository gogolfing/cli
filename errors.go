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

var ErrInvalidParameter = fmt.Errorf("invalid parameter")

type ErrRequiredParameterNotSet struct {
	Name string
	Many bool
}

func (e *ErrRequiredParameterNotSet) Error() string {
	return fmt.Sprintf("required parameter %v not set", FormatParameter(&Parameter{Name: e.Name, Many: e.Many}))
}

var ErrUnsuppliedSubCommand = fmt.Errorf("sub_command not supplied")

type ErrUnknownSubCommand string

func (e ErrUnknownSubCommand) Error() string {
	return fmt.Sprintf("unknown sub_command %q", string(e))
}

type ErrParsingGlobalArgs error

type ErrParsingSubCommand error

type ErrFlagsAfterParameters string

func (e ErrFlagsAfterParameters) Error() string {
	//TODO change this message.
	return fmt.Sprintf("flags present after parameters: %v", string(e))
}

type ErrExecutingSubCommand error

func IsExecutionError(err error) bool {
	return err.(ErrExecutingSubCommand) != nil
}
