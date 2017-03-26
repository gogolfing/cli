package subcommand

import (
	"errors"
	"fmt"
	"testing"
)

func TestSomething(t *testing.T) {
	if errExecute != errExecute {
		t.Fatal()
	}
}

func TestUnknownSubCommandError_Error(t *testing.T) {
	err := UnknownSubCommandError("this is an unknown sub-command")

	if result := err.Error(); result != `unknown sub_command "this is an unknown sub-command"` {
		t.Fail()
	}
}

func TestParsingGlobalArgsError_Error(t *testing.T) {
	err := &ParsingGlobalArgsError{errors.New(t.Name())}
	if err.Error() != t.Name() {
		t.Fail()
	}
}

func TestParsingSubCommandError_Error(t *testing.T) {
	err := &ParsingSubCommandError{errors.New(t.Name())}
	if err.Error() != t.Name() {
		t.Fail()
	}
}

func TestExecutingSubCommandError_Error(t *testing.T) {
	err := &ExecutingSubCommandError{errors.New(t.Name())}
	if err.Error() != t.Name() {
		t.Fail()
	}
}

func TestIsExecutionError(t *testing.T) {
	err := fmt.Errorf("not an execution error")
	if IsExecutionError(err) {
		t.Fail()
	}

	err = &ExecutingSubCommandError{fmt.Errorf("execution error")}
	if !IsExecutionError(err) {
		t.Fail()
	}
}
