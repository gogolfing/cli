package cli

import (
	"fmt"
	"testing"
)

func TestFormatParameter(t *testing.T) {
	tests := []struct {
		p      *Parameter
		result string
	}{
		{
			&Parameter{Name: "one", Optional: false, Many: false},
			"<ONE>",
		},
		{
			&Parameter{Name: "one", Optional: false, Many: true},
			"<ONE...>",
		},
		{
			&Parameter{Name: "one", Optional: true, Many: false},
			"[ONE]",
		},
		{
			&Parameter{Name: "one", Optional: true, Many: true},
			"[ONE...]",
		},
	}

	for i, test := range tests {
		result := FormatParameter(test.p)
		if result != test.result {
			t.Errorf("%v: result = %v WANT %v", i, result, test.result)
		}
	}
}

func testFormatParameter(p *Parameter) string {
	return fmt.Sprintf("FormattedParameter(%v,%v,%v)", p.Name, p.Optional, p.Many)
}

func TestFormatParameters(t *testing.T) {
	params := []*Parameter{
		{Name: "one", Optional: true, Many: false},
		{Name: "two", Optional: false, Many: true},
	}

	result := FormatParameters(params, testFormatParameter)

	want := "FormattedParameter(one,true,false) FormattedParameter(two,false,true)"

	if result != want {
		t.Fail()
	}
}
