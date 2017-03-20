package cli

import (
	"context"
	"flag"
	"io"
)

//SubCommandStruct is a struct type implementation of SubCommand where each field
//represents the value to delegate to in SubCommand methods.
//Nil interface and func values are not used and the zero values are used for
//return values if not supplied.
type SubCommandStruct struct {
	//NameValue is returned from SubCommand's Name() method.
	NameValue string

	//AliasesValue is returned from SubCommand's Aliases() method.
	AliasesValue []string

	//SynopsisValue is returned from SubCommand's Synopsis() method.
	SynopsisValue string

	//DescriptionValue is returned from SubCommand's Description() method.
	DescriptionValue string

	//FlagSetter is used as the SubCommand's implementation for SetFlags if not nil.
	FlagSetter

	//ParameterUsageValue is used as the SubCommand's implementation if not nil.
	ParameterUsageValue func() ([]*Parameter, string)

	//SetParametersValue is used as the SubCommand's implementation if not nil.
	SetParametersValue func([]string) error

	//ExecuteValue is used as the SubCommand's implementation if not nil.
	ExecuteValue func(context.Context, io.Writer, io.Writer) error
}

//Name returns scs.NameValue.
func (scs *SubCommandStruct) Name() string {
	return scs.NameValue
}

//Aliases returns scs.AliasesValue.
func (scs *SubCommandStruct) Aliases() []string {
	return scs.AliasesValue
}

//Synopsis returns scs.SynopsisValue
func (scs *SubCommandStruct) Synopsis() string {
	return scs.SynopsisValue
}

//Description returns scs.DescriptionValue.
func (scs *SubCommandStruct) Description() string {
	return scs.DescriptionValue
}

//SetFlags delegates to scs.FlagSetter if the field is not nil.
func (scs *SubCommandStruct) SetFlags(f *flag.FlagSet) {
	if scs.FlagSetter != nil {
		scs.FlagSetter.SetFlags(f)
	}
}

//ParameterUsage calls and returns the results from scs.ParameterUsageValue()
//if the field is not nil.
//Otherwise, nil and the empty string are returned.
func (scs *SubCommandStruct) ParameterUsage() ([]*Parameter, string) {
	if scs.ParameterUsageValue != nil {
		return scs.ParameterUsageValue()
	}
	return nil, ""
}

//SetParameters calls and returns the result from scs.SetParametersValue(params)
//if the field is not nil.
//Otherwise, nil is returned.
func (scs *SubCommandStruct) SetParameters(params []string) error {
	if scs.SetParametersValue != nil {
		return scs.SetParametersValue(params)
	}
	return nil
}

//Execute calls and returns the result from scs.ExecuteValue(ctx, out, outErr)
//if the field is not nil.
//Otherwise, nil is returned.
func (scs *SubCommandStruct) Execute(ctx context.Context, out, outErr io.Writer) error {
	if scs.ExecuteValue != nil {
		return scs.ExecuteValue(ctx, out, outErr)
	}
	return nil
}
