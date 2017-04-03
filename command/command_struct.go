package command

import (
	"context"
	"flag"
	"io"

	"github.com/gogolfing/cli"
)

//CommandStruct is a struct type implementation of Command where each field
//represents the value to delegate to in Command methods.
//Nil interface and func values are not used and the zero values are used for
//return values if not supplied.
type CommandStruct struct {
	//DescriptionValue is returned from Command's Description() method.
	DescriptionValue string

	//FlagSetter is used as the Command's implementation for SetFlags if not nil.
	cli.FlagSetter

	//ParameterSetter is used as the Command's implementation for ParameterSetter's
	//methods if not nil.
	cli.ParameterSetter

	//ExecuteValue is used as the Command's implementation if not nil.
	ExecuteValue func(context.Context, io.Reader, io.Writer, io.Writer) error
}

//Description returns cs.DescriptionValue.
func (cs *CommandStruct) Description() string {
	return cs.DescriptionValue
}

//SetFlags delegates to cs.FlagSetter if the field is not nil.
func (cs *CommandStruct) SetFlags(f *flag.FlagSet) {
	if cs.FlagSetter != nil {
		cs.FlagSetter.SetFlags(f)
	}
}

//ParameterUsage delegates to cs.ParameterSetter if the field is not nil.
//Otherwise, it returns nil and the empty string.
func (cs *CommandStruct) ParameterUsage() ([]*cli.Parameter, string) {
	if cs.ParameterSetter != nil {
		return cs.ParameterSetter.ParameterUsage()
	}
	return nil, ""
}

//SetParameters delegates to cs.ParameterSetter if the field is not nil.
//Otherwise it returns nil.
func (cs *CommandStruct) SetParameters(params []string) error {
	if cs.ParameterSetter != nil {
		return cs.ParameterSetter.SetParameters(params)
	}
	return nil
}

//Execute calls and returns the result from cs.ExecuteValue(ctx, out, outErr)
//if the field is not nil.
//Otherwise, it returns nil.
func (cs *CommandStruct) Execute(ctx context.Context, in io.Reader, out, outErr io.Writer) error {
	if cs.ExecuteValue != nil {
		return cs.ExecuteValue(ctx, in, out, outErr)
	}
	return nil
}
