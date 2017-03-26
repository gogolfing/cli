package command

import (
	"errors"
	"testing"
)

func TestParsingCommandError(t *testing.T) {
	err := &ParsingCommandError{errors.New(t.Name())}
	if err.Error() != t.Name() {
		t.Fail()
	}
}

func TestExecutingCommandError(t *testing.T) {
	err := &ExecutingCommandError{errors.New(t.Name())}
	if err.Error() != t.Name() {
		t.Fail()
	}
}

func TestIsExecutionError(t *testing.T) {
	err := errors.New("not an executiong error")
	if IsExecutionError(err) {
		t.Fail()
	}

	err = &ExecutingCommandError{errors.New(t.Name())}
	if !IsExecutionError(err) {
		t.Fail()
	}
}
