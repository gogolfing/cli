package cli

import (
	"bytes"
	"flag"
	"io/ioutil"
	"sort"
	"strings"
)

//DoubleMinus is the argument to determine if the flag package has stopped parsing
//after seeing this argument.
const DoubleMinus = "--"

//Output values that affect error and help output.
const (
	Usage = "usage:"

	ParameterName  = "parameter"
	ParametersName = "parameters"

	ArgumentSeparator = " | "
)

//FlagSetter allows implementations to receive values from flag.FlagSets while
//argument parsing occurs.
//Implementations should not retain references to f.
type FlagSetter interface {
	SetFlags(f *flag.FlagSet)
}

//FormatArgument formats an argument's name given whether or not it is optional
//or multiple values are allowed.
func FormatArgument(name string, optional, many bool) string {
	result := name
	if many {
		result += "..."
	}
	if optional {
		result = "[" + result + "]"
	} else {
		result = "<" + result + ">"
	}
	return result
}

//NewFlagSet creates a new flag.FlagSet with name and flag.ContinueOnError and
//calls fs with it if fs is not nil. If fs is nil, then a new, empty flag.FlagSet
//is returned.
func NewFlagSet(name string, fs FlagSetter) *flag.FlagSet {
	f := flag.NewFlagSet(name, flag.ContinueOnError)
	f.Usage = func() {}
	f.SetOutput(ioutil.Discard)
	if fs != nil {
		fs.SetFlags(f)
	}
	return f
}

//CountFlags returns the total number of flags (set or unset) in f.
func CountFlags(f *flag.FlagSet) int {
	count := 0
	f.VisitAll(func(_ *flag.Flag) {
		count++
	})
	return count
}

//GetFlagSetDefaults returns the result of f.PrintDefaults() with the optionally
//trailing "\n" removed.
func GetFlagSetDefaults(f *flag.FlagSet) string {
	out := bytes.NewBuffer([]byte{})
	f.SetOutput(out)
	f.PrintDefaults()
	return strings.TrimRight(out.String(), "\n")
}

//GetJoinedNameSortedAliases return the name followed by the sorted aliases joined
//by ", ".
func GetJoinedNameSortedAliases(name string, aliases []string) string {
	toSort := make([]string, len(aliases))
	copy(toSort, aliases)
	sort.Strings(toSort)
	all := append([]string{name}, toSort...)
	return strings.Join(all, ", ")
}
