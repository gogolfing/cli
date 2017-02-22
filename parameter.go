package cli

import "strings"

type Parameter struct {
	Name     string
	Optional bool
	Many     bool
}

func FormatParameters(params []*Parameter, format func(p *Parameter) string) string {
	return ""
}

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

var FormatParameterName = strings.ToUpper
