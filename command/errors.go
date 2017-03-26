package command

//ParsingCommandError is an error wrapper denoting command argument parsing failed.
type ParsingCommandError struct {
	Err error
}

//Error returns e.Err.Error().
func (e *ParsingCommandError) Error() string {
	return e.Err.Error()
}

//ExecutingCommandError is an error wrapper denoting command execution failed.
type ExecutingCommandError struct {
	Err error
}

//Err returns e.Err.Error().
func (e *ExecutingCommandError) Error() string {
	return e.Err.Error()
}

//IsExecutionError returns whether or not err is an ExecutingcommandError.
func IsExecutionError(err error) bool {
	_, ok := err.(*ExecutingCommandError)
	return ok
}
