package cli

import (
	"flag"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

func TestParseArgumentsInterspersed(t *testing.T) {
	tests := []struct {
		f      *flag.FlagSet
		args   []string
		params []string
		err    error
	}{
		{
			func() *flag.FlagSet {
				f := newFlagSet("")
				return f
			}(),
			[]string{},
			[]string{},
			nil,
		},
		{
			func() *flag.FlagSet {
				f := newFlagSet("")
				return f
			}(),
			[]string{""},
			[]string{""},
			nil,
		},
		{
			func() *flag.FlagSet {
				f := newFlagSet("")
				return f
			}(),
			[]string{"-h"},
			[]string{},
			flag.ErrHelp,
		},
		{
			func() *flag.FlagSet {
				f := newFlagSet("")
				f.Int("a", 0, "")
				f.Bool("b", false, "")
				return f
			}(),
			strings.Fields("-a 10 hello -b world"),
			[]string{"hello", "world"},
			nil,
		},
		{
			func() *flag.FlagSet {
				f := newFlagSet("")
				f.Int("a", 0, "")
				f.Bool("b", false, "")
				return f
			}(),
			strings.Fields("hello world -a 10"),
			[]string{"hello", "world"},
			nil,
		},
		{
			func() *flag.FlagSet {
				f := newFlagSet("")
				f.Int("a", 0, "")
				f.Bool("b", false, "")
				return f
			}(),
			strings.Fields("-a 10 hello"),
			[]string{"hello"},
			nil,
		},
		{
			func() *flag.FlagSet {
				f := newFlagSet("")
				return f
			}(),
			strings.Fields("--"),
			[]string{},
			nil,
		},
		{
			func() *flag.FlagSet {
				f := newFlagSet("")
				return f
			}(),
			strings.Fields("-- -a 10 hello"),
			[]string{"-a", "10", "hello"},
			nil,
		},
		{
			func() *flag.FlagSet {
				f := newFlagSet("")
				f.Int("a", 0, "")
				f.Bool("b", false, "")
				return f
			}(),
			strings.Fields("-a 10 -- -b hello"),
			[]string{"-b", "hello"},
			nil,
		},
	}

	for i, test := range tests {
		params, err := ParseArgumentsInterspersed(test.f, test.args)

		if (len(params) != 0 || len(test.params) != 0) && !reflect.DeepEqual(params, test.params) {
			t.Errorf("%v: ParseArgumentsInterspersed() params = %v WANT %v", i, params, test.params)
		}

		if !reflect.DeepEqual(err, test.err) {
			t.Errorf("%v: ParseArgumentsInterspersed() err = %v WANT %v", i, err, test.err)
		}
	}
}

func newFlagSet(name string) *flag.FlagSet {
	f := flag.NewFlagSet(name, flag.ContinueOnError)
	f.Usage = func() {}
	f.SetOutput(ioutil.Discard)
	return f
}
