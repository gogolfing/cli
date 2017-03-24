package subcommand

import "fmt"

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
	Err error
}

//Error return e.Err.Error().
func (e *ParsingGlobalArgsError) Error() string {
	return e.Err.Error()
}

//ParsingSubCommandError is an error wrapper denoting sub-command argument parsing
//failed.
type ParsingSubCommandError struct {
	Err error
}

//Error return e.Err.Error().
func (e *ParsingSubCommandError) Error() string {
	return e.Err.Error()
}

//ExecutingSubCommandError is an error wrapper denoting that executing a sub-command
//failed.
type ExecutingSubCommandError struct {
	Err error
}

//Error return e.Err.Error().
func (e *ExecutingSubCommandError) Error() string {
	return e.Err.Error()
}

//IsExecutionError returns whether or not err is an ExecutingSubCommandError error.
func IsExecutionError(err error) bool {
	switch err.(type) {
	case *ExecutingSubCommandError:
		return true
	}
	return false
}
