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

//ErrTooManyParameters is a sentinel value that clients can use to signal that
//too many parameters were provided to a ParameterSetter.
var ErrTooManyParameters = fmt.Errorf("too many parameters")
