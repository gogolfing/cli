package cli

import (
	"fmt"
	"testing"
)

func TestExitStatusError_Error(t *testing.T) {
	err := &ExitStatusError{
		Code: 1,
		Err:  fmt.Errorf("error"),
	}

	if err.Error() != "error" {
		t.Fatal()
	}
}

func TestRequiredParameterNotSetError_Error(t *testing.T) {
	err := &RequiredParameterNotSetError{
		Name: "name",
	}
	if err.Error() != "required parameter name not set" {
		t.Fail()
	}

	err = &RequiredParameterNotSetError{
		Formatted: "formatted",
	}
	if err.Error() != "required parameter formatted not set" {
		t.Fail()
	}
}
