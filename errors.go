package cli

import (
	"errors"
	"fmt"
)

//ExitStatusError allows wrapping together an exit status with an error providing
//that exit code.
type ExitStatusError struct {
	//Code is the desired exit status.
	Code int

	//Err is the wrapped error.
	Err error
}

//Error is the error implementation for e.
//It returns e.Err.Error()
func (e *ExitStatusError) Error() string {
	return e.Err.Error()
}

//ErrInvalidParameters is a generic error for invalid parameters being set.
//Note that this error message will not be printed to output, it is simply a sentinel
//value.
var ErrInvalidParameters = errors.New("invalid parameters")

//RequiredParameterNotSetError is an error that denotes a parameter was not supplied
//in the arguments but is required to be present.
//If Formatted is not empty, then it is used in Error(). Otherwise, Name is used.
//It is up to client code to set Formatted if that output is desired.
type RequiredParameterNotSetError struct {
	//Name is the name of the Parameter.
	Name string

	//Many is the Many option of the Parameter.
	Many bool

	//Formatted may be provided to override output of Error().
	Formatted string
}

//Error provides the error implementation.
//It returns fmt.Sprintf("required %s %s not set", ParameterName, <value>) where
//<value> is e.Formatted if not empty or e.Name otherwise.
func (e *RequiredParameterNotSetError) Error() string {
	value := e.Name
	if len(e.Formatted) != 0 {
		value = e.Formatted
	}
	return fmt.Sprintf(
		"required %s %s not set",
		ParameterName,
		value,
	)
}

//ErrTooManyParameters is a sentinel value that clients can use to signal that
//too many parameters were provided to a ParameterSetter.
var ErrTooManyParameters = fmt.Errorf("too many parameters")
