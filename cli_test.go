package cli

import "flag"

type FlagSetterFunc func(*flag.FlagSet)

func (fsf FlagSetterFunc) SetFlags(f *flag.FlagSet) {
	fsf(f)
}

func NewStringsFlagSetter(names ...string) FlagSetter {
	return FlagSetterFunc(func(f *flag.FlagSet) {
		for _, name := range names {
			f.String(name, name+"_default", name+"_usage")
		}
	})
}
