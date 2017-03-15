package cli

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

const (
	SimpleUsage        = "usage: command <sub_command> [parameters...] [sub_command_options...]\n"
	SimpleGlobalsUsage = "usage: command [global_options...] <sub_command> [parameters...] [sub_command_options...]\n"
)

var errExecute = errors.New("error executing")

func TestSubCommander_RegisterHelp_RegistersWithNameAndAlaises(t *testing.T) {
	sc := &SubCommander{}

	sc.RegisterHelp("help", "", "", "h")

	if sc.names["help"] == nil {
		t.Fatalf("help should be registered")
	}
	if sc.aliases["h"] == nil {
		t.Fatalf("h should be registerd")
	}
}

func TestSubCommander_RegisterListSubCommands_RegistersWithNameAndAliases(t *testing.T) {
	sc := &SubCommander{}

	sc.RegisterListSubCommands("list", "", "", "ls")

	if sc.names["list"] == nil {
		t.Fatalf("list should be registered")
	}
	if sc.aliases["ls"] == nil {
		t.Fatalf("ls should be registered")
	}
}

func TestSubCommander_Register_RegistersSubCommandsNameAndAliases(t *testing.T) {
	sc := &SubCommander{}

	subCommand := &SubCommandStruct{
		NameValue:    "name",
		AliasesValue: []string{"a", "b"},
	}
	sc.Register(subCommand)

	if sc.names["name"] != subCommand {
		t.Fatalf("name should be registered with subCommand")
	}
}

func TestSubCommander_Execute_CallsExecuteContextOutCorrectly(t *testing.T) {
	oldOut, oldErr := os.Stdout, os.Stderr
	defer func() {
		os.Stdout, os.Stderr = oldOut, oldErr
	}()

	tempDir, _ := ioutil.TempDir(".", "")
	defer os.RemoveAll(tempDir)
	out, _ := ioutil.TempFile(tempDir, "")
	outErr, _ := ioutil.TempFile(tempDir, "")
	os.Stdout, os.Stderr = out, outErr

	sc := &SubCommander{}

	subCommand := &SubCommandStruct{
		NameValue:    "a",
		ExecuteValue: NewExecuteFunc("out", "outErr", errExecute),
	}
	sc.Register(subCommand)

	err := sc.Execute(strings.Fields("a"))

	out.Close()
	outErr.Close()

	if outBytes, _ := ioutil.ReadFile(os.Stdout.Name()); string(outBytes) != "out" {
		t.Fatalf("out = %v WANT %s", string(outBytes), "out")
	}
	if outErrBytes, _ := ioutil.ReadFile(os.Stderr.Name()); string(outErrBytes) != "outErr" {
		t.Fatalf("outErr = %s WANT %s", outErrBytes, "outErr")
	}
	if err.Error() != errExecute.Error() {
		t.Fatalf("err = %v WANT %v", err, errExecute)
	}
}

func TestSubCommander_ExecuteContextOut_GlobalFlagParsingError_Help(t *testing.T) {
	sct := &SubCommanderTest{
		Args:         strings.Fields("-h"),
		OutErrString: SimpleUsage,
		Err:          &ParsingGlobalArgsError{flag.ErrHelp},
	}

	testSubCommanderTest(t, sct)
}

func TestSubCommander_ExecuteContextOut_GlobalFlagParsingError_OtherError(t *testing.T) {
	errString := "flag provided but not defined: -other"

	fs := NewStringsFlagSetter("value")
	sct := &SubCommanderTest{
		SubCommander: &SubCommander{
			GlobalFlags: fs,
		},
		Args:         strings.Fields("-other 1234"),
		OutErrString: errString + "\n\n" + SimpleGlobalsUsage + "\n" + GlobalOptionsName + ":\n" + getFlagSetterDefaults(fs),
		Err:          &ParsingGlobalArgsError{errors.New(errString)},
	}

	testSubCommanderTest(t, sct)
}

func TestSubCommander_ExecuteContextOut_UnsuppliedSubCommandError(t *testing.T) {
	prefix := ErrUnsuppliedSubCommand.Error() + "\n\n"

	tests := []*SubCommanderTest{
		{
			Args:         nil,
			OutErrString: prefix + SimpleUsage,
			Err:          ErrUnsuppliedSubCommand,
		},
		{
			Args:         []string{},
			OutErrString: prefix + SimpleUsage,
			Err:          ErrUnsuppliedSubCommand,
		},
		func() *SubCommanderTest {
			fs := NewStringsFlagSetter("value")
			return &SubCommanderTest{
				SubCommander: &SubCommander{
					GlobalFlags: fs,
				},
				Args:         []string{"-value", "1234"},
				OutErrString: prefix + SimpleGlobalsUsage,
				Err:          ErrUnsuppliedSubCommand,
			}
		}(),
	}

	testSubCommanderTests(t, tests)
}

func TestSubCommander_ExecuteContextOut_UnsuppliedSubCommandError_PrintsAvailableSubCommandsCorrectly(t *testing.T) {
	prefix := ErrUnsuppliedSubCommand.Error() + "\n\n"
	subCommandListing :=
		`  a               command a
  b, b1, b2       command b`

	sct := &SubCommanderTest{
		SubCommands: []SubCommand{
			&SubCommandStruct{
				NameValue:     "b",
				AliasesValue:  []string{"b2", "b1"},
				SynopsisValue: "command b",
			},
			&SubCommandStruct{
				NameValue:     "a",
				SynopsisValue: "command a",
			},
		},
		Args:         nil,
		OutErrString: prefix + SimpleUsage + "\n" + SubCommandsName + ":\n" + subCommandListing + "\n",
		Err:          ErrUnsuppliedSubCommand,
	}

	testSubCommanderTest(t, sct)
}

func TestSubCommander_ExecuteContextOut_UnsuppliedSubCommandError_PrintsAvailableSubCommandsWithWideOutput(t *testing.T) {
	prefix := ErrUnsuppliedSubCommand.Error() + "\n\n"
	subCommandListing :=
		"  a, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9    command a"

	sct := &SubCommanderTest{
		SubCommands: []SubCommand{
			&SubCommandStruct{
				NameValue:     "a",
				AliasesValue:  []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"},
				SynopsisValue: "command a",
			},
		},
		Args:         nil,
		OutErrString: prefix + SimpleUsage + "\n" + SubCommandsName + ":\n" + subCommandListing + "\n",
		Err:          ErrUnsuppliedSubCommand,
	}

	testSubCommanderTest(t, sct)
}

func TestSubCommander_ExecuteContextOut_UnknownSubCommandError(t *testing.T) {
	prefix := UnknownSubCommandError("foo").Error() + "\n\n"

	sct := &SubCommanderTest{
		Args:         strings.Fields("foo"),
		OutErrString: prefix + SimpleUsage,
		Err:          UnknownSubCommandError("foo"),
	}

	testSubCommanderTest(t, sct)
}

func TestSubCommander_ExecuteContextOut_ParsingSubCommandError_FlagsFirstError(t *testing.T) {
	err := fmt.Errorf("this is my error")
	// err := fmt.Errorf(t.Name())

	sct := &SubCommanderTest{
		SubCommander: &SubCommander{
			ParameterFlagMode: ModeFlagsFirst,
		},
		SubCommands: []SubCommand{
			&SubCommandStruct{
				NameValue: "sub",
				SetParametersValue: func(params []string) error {
					if !reflect.DeepEqual(params, strings.Fields("param1 -f 2")) {
						t.Fatal("wrong parameters")
					}
					return err
				},
				ExecuteValue: NewExecuteFunc("out", "outErr", errExecute),
			},
		},
		Args:         strings.Fields("sub param1 -f 2"),
		OutErrString: "",
		Err:          &ParsingSubCommandError{err},
	}

	testSubCommanderTest(t, sct)
}

func TestSubCommander_ExecuteContextOut_ParsingSubCommandError_ParametersFirstError(t *testing.T) {
}

func TestSubCommander_ExecuteContextOut_ParsingSubCommandError_InterspersedError(t *testing.T) {
}

func TestSubCommander_ExecuteContextOut_ParsingSubCommandError_SettingParametersError(t *testing.T) {
}

//TODO works correctly with disallow globals option set
//TODO erros with disallow globals option set
//TODO works correctly with all three modes
//TODO works correctly with alias
//TODO works correctly with global option in sub command args
//TODO help works
//TODO list works

func NewExecuteFunc(out, outErr string, err error) func(context.Context, io.Writer, io.Writer) error {
	f := func(_ context.Context, outW, outErrW io.Writer) error {
		fmt.Fprint(outW, out)
		fmt.Fprint(outErrW, outErr)
		return err
	}
	return f
}

type SubCommanderTest struct {
	*SubCommander

	//CommandName is set here as a convenience to setting it on SubCommander.
	CommandName string

	SubCommands  []SubCommand
	RegisterHelp bool
	RegisterList bool

	Context context.Context
	Args    []string

	OutString    string
	OutErrString string
	Err          error
}

func testSubCommanderTests(t *testing.T, tests []*SubCommanderTest) {
	for i, test := range tests {
		testSubCommanderTest(t, test, i)
	}
}

func testSubCommanderTest(t *testing.T, sct *SubCommanderTest, tags ...interface{}) {
	prefix := fmt.Sprintf("%v: ", t.Name())
	if len(tags) > 0 {
		prefix += fmt.Sprintf("%v: ", tags[0])
	}
	prefix += "sc.ExecuteContextOut()"

	sc := sct.SubCommander
	if sc == nil {
		sc = &SubCommander{}
	}

	if sc.CommandName == "" {
		if sct.CommandName != "" {
			sc.CommandName = sct.CommandName
		} else {
			sc.CommandName = "command"
		}
	}

	for _, subCommand := range sct.SubCommands {
		sc.Register(subCommand)
	}
	if sct.RegisterHelp {
		sc.RegisterHelp("help", "", "")
	}
	if sct.RegisterList {
		sc.RegisterListSubCommands("list", "", "")
	}

	if sct.Context == nil {
		sct.Context = context.Background()
	}

	out, outErr, err := executeContextOut(sc, sct.Context, sct.Args)

	outString := out.String()
	outErrString := outErr.String()

	if outString != sct.OutString {
		t.Errorf(
			"%v: out = %v WANT %v",
			t.Name(),
			outString,
			sct.OutString,
		)
	}

	if outErrString != sct.OutErrString {
		t.Errorf(
			"%v: outErr = %v WANT %v",
			prefix,
			outErrString,
			sct.OutErrString,
		)
	}

	if !reflect.DeepEqual(err, sct.Err) {
		t.Errorf(
			"%v: err = %v WANT %v",
			prefix,
			err,
			sct.Err,
		)
	}
}

func executeContextOut(sc *SubCommander, ctx context.Context, args []string) (*bytes.Buffer, *bytes.Buffer, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	out, outErr := newOutputs()

	err := sc.ExecuteContextOut(ctx, args, out, outErr)

	return out, outErr, err
}

func newOutputs() (*bytes.Buffer, *bytes.Buffer) {
	return bytes.NewBuffer([]byte{}), bytes.NewBuffer([]byte{})
}

func getFlagSetterDefaults(fs FlagSetter) string {
	f := flag.NewFlagSet("", flag.ContinueOnError)
	out := bytes.NewBuffer([]byte{})
	fs.SetFlags(f)
	f.SetOutput(out)
	f.PrintDefaults()
	return out.String()
}