package cli

import "testing"

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
