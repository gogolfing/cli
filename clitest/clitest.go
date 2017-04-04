//Package clitest provides useful types and functions for testing cli subpackages.
package clitest

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/gogolfing/cli"
)

//FlagSetterFunc is a function implementation for a FlagSetter.
type FlagSetterFunc func(*flag.FlagSet)

//SetFlags calls fsf(f).
func (fsf FlagSetterFunc) SetFlags(f *flag.FlagSet) {
	fsf(f)
}

//NewStringsFlagSetter sets each name in names as a string in the returned FlagSetter.
//The default value is (name+"_default") and the usage value is (name+"_usage").
func NewStringsFlagSetter(names ...string) cli.FlagSetter {
	return FlagSetterFunc(func(f *flag.FlagSet) {
		for _, name := range names {
			f.String(name, name+"_default", name+"_usage")
		}
	})
}

//SimpleFlagSetter is a cli.FlagSetter that allows client code to set flags and
//use those set values at a later time.
type SimpleFlagSetter struct {
	//Suffix is the suffix added to the names of the flags in SetFlags.
	Suffix string

	//Int is where a flag with name ("int"+Suffix) is set.
	Int int

	//String is where a flag with name ("string"+Suffix) is set.
	String string

	//Bool is where a flag with name ("bool"+Suffix) is set.
	Bool bool
}

//SetFlags is the FlagSetter implementation for sfs.
func (sfs *SimpleFlagSetter) SetFlags(f *flag.FlagSet) {
	f.IntVar(&sfs.Int, "int"+sfs.Suffix, sfs.Int, "int_usage")
	f.StringVar(&sfs.String, "string"+sfs.Suffix, sfs.String, "string_usage")
	f.BoolVar(&sfs.Bool, "bool"+sfs.Suffix, sfs.Bool, "bool_usage")
}

//NewExecuteFunc returns a function that matches the command.Command.Execute and
//subcommand.SubCommand.Execute signatures.
//
//The resulting function calls ioutil.ReadAll on in and writes out and outErr to
//their respective io.Writers.
//It returns err.
func NewExecuteFunc(out, outErr string, err error) func(context.Context, io.Reader, io.Writer, io.Writer) error {
	f := func(_ context.Context, inR io.Reader, outW, outErrW io.Writer) error {
		ioutil.ReadAll(inR)
		fmt.Fprint(outW, out)
		fmt.Fprint(outErrW, outErr)
		return err
	}
	return f
}

//NewOutputs two bytes.NewBuffer()s each with an empty byte{} buffer.
func NewOutputs() (*bytes.Buffer, *bytes.Buffer) {
	return bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{})
}

//GetFlagSetterDefaults returns the result of flag.FlagSet.PrintDefaults on a
//flag.FlagSet being called on fs.
func GetFlagSetterDefaults(fs cli.FlagSetter) string {
	f := flag.NewFlagSet("", flag.ContinueOnError)
	out := bytes.NewBuffer([]byte{})
	fs.SetFlags(f)
	f.SetOutput(out)
	f.PrintDefaults()
	return strings.TrimRight(out.String(), "\n")
}

//ParameterSetterStruct is a struct implementation for cli.ParameterSetter.
type ParameterSetterStruct struct {
	//ParameterUsageValue is used as the implementation if not nil.
	ParameterUsageValue func() ([]*cli.Parameter, string)

	//SetParametersValue is used as the implementation if not nil.
	SetParametersValue func([]string) error
}

//ParameterUsage calls pss.ParameterUsageValue if it is not nil.
//Otherwise returns nil and the empty string.
func (pss *ParameterSetterStruct) ParameterUsage() ([]*cli.Parameter, string) {
	if pss.ParameterUsageValue != nil {
		return pss.ParameterUsageValue()
	}
	return nil, ""
}

//SetParameters delegates to pss.SetParametersValue if it is not nil.
//Otherwise returns nil.
func (pss *ParameterSetterStruct) SetParameters(params []string) error {
	if pss.SetParametersValue != nil {
		return pss.SetParametersValue(params)
	}
	return nil
}
