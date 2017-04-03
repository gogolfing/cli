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

type FlagSetterFunc func(*flag.FlagSet)

func (fsf FlagSetterFunc) SetFlags(f *flag.FlagSet) {
	fsf(f)
}

func NewStringsFlagSetter(names ...string) cli.FlagSetter {
	return FlagSetterFunc(func(f *flag.FlagSet) {
		for _, name := range names {
			f.String(name, name+"_default", name+"_usage")
		}
	})
}

type SimpleFlagSetter struct {
	Suffix string

	Int    int
	String string
	Bool   bool
}

func (sfs *SimpleFlagSetter) SetFlags(f *flag.FlagSet) {
	f.IntVar(&sfs.Int, "int"+sfs.Suffix, sfs.Int, "int_usage")
	f.StringVar(&sfs.String, "string"+sfs.Suffix, sfs.String, "string_usage")
	f.BoolVar(&sfs.Bool, "bool"+sfs.Suffix, sfs.Bool, "bool_usage")
}

func NewExecuteFunc(out, outErr string, err error) func(context.Context, io.Reader, io.Writer, io.Writer) error {
	f := func(_ context.Context, inR io.Reader, outW, outErrW io.Writer) error {
		ioutil.ReadAll(inR)
		fmt.Fprint(outW, out)
		fmt.Fprint(outErrW, outErr)
		return err
	}
	return f
}

func NewOutputs() (*bytes.Buffer, *bytes.Buffer) {
	return bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{})
}

func GetFlagSetterDefaults(fs cli.FlagSetter) string {
	f := flag.NewFlagSet("", flag.ContinueOnError)
	out := bytes.NewBuffer([]byte{})
	fs.SetFlags(f)
	f.SetOutput(out)
	f.PrintDefaults()
	return strings.TrimRight(out.String(), "\n")
}

type ParameterSetterStruct struct {
	ParameterUsageValue func() ([]*cli.Parameter, string)
	SetParametersValue  func([]string) error
}

func (pss *ParameterSetterStruct) ParameterUsage() ([]*cli.Parameter, string) {
	if pss.ParameterUsageValue != nil {
		return pss.ParameterUsageValue()
	}
	return nil, ""
}

func (pss *ParameterSetterStruct) SetParameters(params []string) error {
	if pss.SetParametersValue != nil {
		return pss.SetParametersValue(params)
	}
	return nil
}
