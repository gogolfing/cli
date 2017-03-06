package cli

import "strings"

//Parameter is a value struct for a parameter in the command line arguments.
type Parameter struct {
	//Name is the name of the Parameter as it will appear in help and error output.
	Name string

	//Optional denotes whether or not the parameter's presence in command line
	//arguments is optional. False for required.
	Optional bool

	//Many denotes whether or not this Parameter can have a variable number of
	//command line arguments for input.
	Many bool
}

//FormatParameters calls format() for each Parameter in params and returns
//the result joined by " ".
func FormatParameters(params []*Parameter, format func(p *Parameter) string) string {
	formats := make([]string, 0, len(params))
	for _, param := range params {
		formats = append(formats, format(param))
	}
	return strings.Join(formats, " ")
}

//FormatParameter returns a string representation of p appropriate for help and
//error output.
//
//This value may be changed to affect the output of this package.
var FormatParameter = func(p *Parameter) string {
	result := FormatParameterName(p.Name)
	if p.Many {
		result += "..."
	}
	if p.Optional {
		result = "[" + result + "]"
	}
	return result
}

//FormatParameterName returns a string representation of a Parameter name appropriate
//for help and error output.
//
//This value my be changed to affect the output of this package.
var FormatParameterName = strings.ToUpper
