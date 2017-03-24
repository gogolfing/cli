package clitest

import (
	"flag"

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
