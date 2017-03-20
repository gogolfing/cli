package cli

import (
	"errors"
	"fmt"
)

//ExitStatusError allows wrapping together an exit status with an error providing
//that exist code.
type ExitStatusError struct {
	//Code is the desired exit status.
	Code int

	//error is the wrapped error.
	error
}

//ErrInvalidParameters is a generic error for invalid parameters being set.
//Note that this error message will not be printed to output, it is simply a sentinel
//value.
var ErrInvalidParameters = errors.New("invalid parameters")

//RequiredParameterNotSetError is an error that denotes a parameter was not supplied
//in the arguments but is required to be present.
type RequiredParameterNotSetError struct {
	Name string
	Many bool
}

//Error provides the error implementation.
func (e *RequiredParameterNotSetError) Error() string {
	return fmt.Sprintf(
		"required %v %v not set",
		ParameterName,
		FormatParameter(&Parameter{Name: e.Name, Many: e.Many}),
	)
}

//ErrUnsuppliedSubCommand is a value error denoting a sub-command was not supplied
//during argument parsing.
//Note that this error message will not be printed to output, it is simply a sentinel
//value.
var ErrUnsuppliedSubCommand = fmt.Errorf("%s not supplied", SubCommandName)

//UnknownSubCommandError is an error denoting the provided sub-command is not registered.
type UnknownSubCommandError string

//Error provides the error implementation.
func (e UnknownSubCommandError) Error() string {
	return fmt.Sprintf("unknown %v %q", SubCommandName, string(e))
}

//ParsingGlobalArgsError is an error wrapper denoting global argument parsing failed.
type ParsingGlobalArgsError struct {
	error
}

//ParsingSubCommandError is an error wrapper denoting sub-command argument parsing
//failed.
type ParsingSubCommandError struct {
	error
}

//ErrTooManyParameters is a sentinel value that clients can use to signal that
//too many parameters were provided to a ParameterSetter.
var ErrTooManyParameters = fmt.Errorf("too many parameters")

//ExecutingSubCommandError is an error wrapper denoting that executing a sub-command
//failed.
type ExecutingSubCommandError struct {
	error
}

//IsExecutionError returns whether or not err is an ExecutingSubCommandError error.
func IsExecutionError(err error) bool {
	switch err.(type) {
	case ExecutingSubCommandError, *ExecutingSubCommandError:
		return true
	}
	return false
}
