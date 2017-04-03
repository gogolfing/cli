package cli

import (
	"flag"
	"fmt"
	"reflect"
	"testing"
)

func TestFormatArgument(t *testing.T) {
	tests := []struct {
		name     string
		optional bool
		many     bool
		result   string
	}{
		{"arg", false, false, "<arg>"},
		{"arg", false, true, "<arg...>"},
		{"arg", true, false, "[arg]"},
		{"arg", true, true, "[arg...]"},
	}

	for i, test := range tests {
		result := FormatArgument(test.name, test.optional, test.many)
		if result != test.result {
			t.Errorf("%v: FormatArgument() = %v WANT %v", i, result, test.result)
		}
	}
}

func TestNewFlagSet(t *testing.T) {
	fs := &MockFlagSetter{}
	f := NewFlagSet("name", fs)

	if fs.calledWith != f {
		t.Fatal()
	}
}

func TestCountFlags(t *testing.T) {
	fs := NewFlagSet("", IntFlagSetter(0))
	if CountFlags(fs) != 0 {
		t.Fatal()
	}

	fs = NewFlagSet("", IntFlagSetter(200))
	if CountFlags(fs) != 200 {
		t.Fatal()
	}
}

func TestGetFlagSetDefaults(t *testing.T) {
	defs := GetFlagSetDefaults(NewFlagSet("", IntFlagSetter(0)))
	if defs != "" {
		t.Fatal(defs)
	}

	defs = GetFlagSetDefaults(NewFlagSet("two", IntFlagSetter(2)))
	want := `  -int1 int
    	int1_usage (default 1)
  -int2 int
    	int2_usage (default 2)`
	if defs != want {
		t.Fatal(defs, want)
	}
}

func TestGetJoinedNameSortedAliases(t *testing.T) {
	aliases := []string{"c", "b", "a"}
	result := GetJoinedNameSortedAliases("d", aliases)
	if result != "d, a, b, c" {
		t.Fatal(result)
	}

	if !reflect.DeepEqual(aliases, []string{"c", "b", "a"}) {
		t.Fatal()
	}
}

type IntFlagSetter int

func (fs IntFlagSetter) SetFlags(f *flag.FlagSet) {
	for i := 1; i <= int(fs); i++ {
		name := fmt.Sprintf("int%d", i)
		f.Int(name, i, name+"_usage")
	}
}

type MockFlagSetter struct {
	calledWith *flag.FlagSet
}

func (fs *MockFlagSetter) SetFlags(f *flag.FlagSet) {
	fs.calledWith = f
}
