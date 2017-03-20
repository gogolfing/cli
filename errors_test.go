package cli

import (
	"fmt"
	"testing"
)

func TestRequiredParameterNotSetError_Error(t *testing.T) {
	oldFormatParameter := FormatParameter
	FormatParameter = testFormatParameter
	defer func() {
		FormatParameter = oldFormatParameter
	}()

	tests := []struct {
		name   string
		many   bool
		result string
	}{
		{"foo", true, "required parameter FormattedParameter(foo,false,true) not set"},
		{"bar", false, "required parameter FormattedParameter(bar,false,false) not set"},
	}

	for i, test := range tests {
		err := &RequiredParameterNotSetError{
			Name: test.name,
			Many: test.many,
		}
		result := err.Error()
		if result != test.result {
			t.Errorf("%v: result = %v WANT %v", i, result, test.result)
		}
	}
}

func TestUnknownSubCommandError_Error(t *testing.T) {
	err := UnknownSubCommandError("this is an unknown sub-command")

	if result := err.Error(); result != `unknown sub_command "this is an unknown sub-command"` {
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

	err = ExecutingSubCommandError{fmt.Errorf("execution error")}
	if !IsExecutionError(err) {
		t.Fail()
	}
}
