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

//ParameterSetter provides the interface for a cli working with command line parameters.
type ParameterSetter interface {
	//ParameterUsage returns the Parameters used and a possible usage string to
	//describe parameters in more detail.
	//These values are used in help and error output.
	ParameterUsage() (params []*Parameter, usage string)

	//SetParameters allows implementations to receive parameter arguments during
	//argument parsing.
	//An error should be returned if values cannot be correctly parsed as parameters
	//expected by the implementation.
	//Implementations should not retain references to values.
	SetParameters(values []string) error
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
func FormatParameter(p *Parameter) string {
	return FormatArgument(FormatParameterName(p.Name), p.Optional, p.Many)
}

//FormatParameterName returns a string representation of a Parameter name appropriate
//for help and error output.
//It returns the result of strings.ToUpper(name).
func FormatParameterName(name string) string {
	return strings.ToUpper(name)
}
