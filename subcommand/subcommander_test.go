package subcommand

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/gogolfing/cli"
	"github.com/gogolfing/cli/clitest"
)

const (
	SimpleUsage        = "usage: command <sub_command> [[sub_command_options | parameters]...]\n"
	SimpleGlobalsUsage = "usage: command [global_options...] <sub_command> [[global_options | sub_command_options | parameters]...]\n"
)

var errExecute = fmt.Errorf("error executing")

func TestSubCommander_RegisterHelp_RegistersWithNameAndAlaises(t *testing.T) {
	sc := &SubCommander{}

	sc.RegisterHelp("help", "", "", "h")

	if sc.names["help"] == nil {
		t.Fatalf("help should be registered")
	}
	if sc.aliases["h"] == nil {
		t.Fatalf("h should be registered")
	}
}

func TestSubCommander_RegisterListSubCommands_RegistersWithNameAndAliases(t *testing.T) {
	sc := &SubCommander{}

	sc.RegisterList("list", "", "", "ls")

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

func TestSubCommander_Execute_CallsExecuteContextCorrectly(t *testing.T) {
	sc := &SubCommander{}

	subCommand := &SubCommandStruct{
		NameValue: "a",
		ExecuteValue: func(_ context.Context, in io.Reader, out, outErr io.Writer) error {
			if in.(*os.File) != os.Stdin {
				t.Fatal("did not call execute with stdin")
			}
			if out.(*os.File) != os.Stdout {
				t.Fatal("did not call execute with stdin")
			}
			if outErr.(*os.File) != os.Stderr {
				t.Fatal("did not call execute with stdin")
			}
			return errExecute
		},
	}
	sc.Register(subCommand)

	err := sc.Execute(strings.Fields("a"))

	if !reflect.DeepEqual(err, &ExecutingSubCommandError{errExecute}) {
		t.Fatalf("err = %v WANT %v", err, errExecute)
	}
}

func TestSubCommander_ExecuteContext_GlobalFlagParsingError_Help(t *testing.T) {
	sct := &SubCommanderTest{
		Args:         strings.Fields("-h"),
		OutErrString: SimpleUsage,
		Err:          &ParsingGlobalArgsError{flag.ErrHelp},
	}

	testSubCommanderTest(t, sct)
}

func TestSubCommander_ExecuteContext_GlobalFlagParsingError_OtherError(t *testing.T) {
	errString := "flag provided but not defined: -other"

	fs := clitest.NewStringsFlagSetter("value")
	sct := &SubCommanderTest{
		SubCommander: &SubCommander{
			GlobalFlags: fs,
		},
		Args:         strings.Fields("-other 1234"),
		OutErrString: errString + "\n\n" + SimpleGlobalsUsage + "\n" + GlobalOptionsName + ":\n" + clitest.GetFlagSetterDefaults(fs) + "\n",
		Err:          &ParsingGlobalArgsError{errors.New(errString)},
	}

	testSubCommanderTest(t, sct)
}

func TestSubCommander_ExecuteContext_UnsuppliedSubCommandError(t *testing.T) {
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
			fs := clitest.NewStringsFlagSetter("value")
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

func TestSubCommander_ExecuteContext_UnsuppliedSubCommandError_PrintsAvailableSubCommandsCorrectly(t *testing.T) {
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

func TestSubCommander_ExecuteContext_UnsuppliedSubCommandError_PrintsAvailableSubCommandsWithWideOutput(t *testing.T) {
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

func TestSubCommander_ExecuteContext_UnknownSubCommandError(t *testing.T) {
	prefix := UnknownSubCommandError("foo").Error() + "\n\n"

	sct := &SubCommanderTest{
		Args:         strings.Fields("foo"),
		OutErrString: prefix + SimpleUsage,
		Err:          UnknownSubCommandError("foo"),
	}

	testSubCommanderTest(t, sct)
}

func TestSubCommander_ExecuteContext_ParsingSubCommandError_FlagErrHelp(t *testing.T) {
	err := flag.ErrHelp

	sct := &SubCommanderTest{
		SubCommands: []SubCommand{
			&SubCommandStruct{
				NameValue: "sub",
			},
		},
		Args:         strings.Fields("sub -h"),
		OutErrString: "sub" + "\n\n" + Usage + " ... sub" + "\n",
		Err:          &ParsingSubCommandError{err},
	}

	testSubCommanderTest(t, sct)
}

func TestSubCommander_ExecuteContext_ParsingSubCommandError_SettingParametersError(t *testing.T) {
	err := fmt.Errorf(t.Name())

	sct := &SubCommanderTest{
		SubCommands: []SubCommand{
			&SubCommandStruct{
				NameValue: "a",
				ParameterSetter: &clitest.ParameterSetterStruct{
					SetParametersValue: func(params []string) error {
						if !reflect.DeepEqual(params, []string{"foo", "bar"}) {
							t.Fatal("wrong parameters")
						}
						return err
					},
					ParameterUsageValue: func() ([]*cli.Parameter, string) {
						return []*cli.Parameter{
							&cli.Parameter{
								Name:     "PV",
								Optional: true,
								Many:     false,
							},
						}, "extra parameter usage"
					},
				},
			},
		},
		Args: strings.Fields("a foo bar"),
		OutErrString: err.Error() + "\n\n" + "usage: ... a [parameters...]" + "\n\n" +
			"parameters: [PV]" + "\n" + "extra parameter usage" + "\n",
		Err: &ParsingSubCommandError{err},
	}

	testSubCommanderTest(t, sct)
}

func TestSubCommander_ExecuteContext_ParsingSubCommandError_SettingParametersError_StillPrintsExtraUsageOnNoParameterSetError(t *testing.T) {
	err := cli.ErrTooManyParameters

	sct := &SubCommanderTest{
		SubCommands: []SubCommand{
			&SubCommandStruct{
				NameValue: "a",
				ParameterSetter: &clitest.ParameterSetterStruct{
					SetParametersValue: func(_ []string) error {
						return err
					},
					ParameterUsageValue: func() ([]*cli.Parameter, string) {
						return nil, "extra parameter usage"
					},
				},
			},
		},
		Args: strings.Fields("a foo bar"),
		OutErrString: err.Error() + "\n\n" + "usage: ... a" + "\n\n" +
			"extra parameter usage" + "\n",
		Err: &ParsingSubCommandError{err},
	}

	testSubCommanderTest(t, sct)
}

func TestSubCommander_ExecuteContext_ParsingSubCommandError_SettingSameSubCommandAndGlobalFlagPanics(t *testing.T) {
	gfs := clitest.NewStringsFlagSetter("foo")
	sfs := clitest.NewStringsFlagSetter("foo")

	sct := &SubCommanderTest{
		SubCommander: &SubCommander{
			GlobalFlags: gfs,
		},
		SubCommands: []SubCommand{
			&SubCommandStruct{
				NameValue:  "sub",
				FlagSetter: sfs,
			},
		},
		Args: strings.Fields("sub"),
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal()
		}
	}()

	testSubCommanderTest(t, sct)
}

func TestSubCommander_ExecuteContext_WorksCorrectlyWithDisallowGlobalOptionsSet(t *testing.T) {
	gfs := clitest.NewStringsFlagSetter("g1")
	sfs := clitest.NewStringsFlagSetter("s1")

	sct := &SubCommanderTest{
		SubCommander: &SubCommander{
			GlobalFlags:                       gfs,
			DisallowGlobalFlagsWithSubCommand: true,
		},
		SubCommands: []SubCommand{
			&SubCommandStruct{
				NameValue:  "sub",
				FlagSetter: sfs,
			},
		},
		Args: strings.Fields("-g1 foo sub -s1 bar"),
		Err:  nil,
	}

	testSubCommanderTest(t, sct)
}

func TestSubCommander_ExecuteContext_ErrorsWithDisallowGlobalsAndGlobalOptionSetAfterSubCommand(t *testing.T) {
	gfs := clitest.NewStringsFlagSetter("g1")
	sfs := clitest.NewStringsFlagSetter("s1")
	err := fmt.Errorf("flag provided but not defined: %v", "-g1")

	sct := &SubCommanderTest{
		SubCommander: &SubCommander{
			GlobalFlags:                       gfs,
			DisallowGlobalFlagsWithSubCommand: true,
		},
		SubCommands: []SubCommand{
			&SubCommandStruct{
				NameValue:  "sub",
				FlagSetter: sfs,
			},
		},
		Args: strings.Fields("sub -g1 foo -s1 bar"),
		OutErrString: err.Error() + "\n\n" + Usage + " ... sub [sub_command_options...]" + "\n\n" +
			SubCommandOptionsName + ":\n" + clitest.GetFlagSetterDefaults(sfs) + "\n",
		Err: &ParsingSubCommandError{err},
	}

	testSubCommanderTest(t, sct)
}

func TestSubCommander_ExecuteContext_WorksCorrectlyWithAlias(t *testing.T) {
	sct := &SubCommanderTest{
		SubCommands: []SubCommand{
			&SubCommandStruct{
				NameValue:    "foo",
				AliasesValue: []string{"bar"},
				ExecuteValue: clitest.NewExecuteFunc("foo bar", "", nil),
			},
		},
		Args:      strings.Fields("bar"),
		OutString: "foo bar",
	}

	testSubCommanderTest(t, sct)
}

//Test to cover Issue #8.
//This ensures that when a global flag is set with a default value, the non-presence
//of the flag in sub-command arguments does not override what was set in the global
//arguments.
func TestSubCommander_ExecuteContext_DefaultSubCommandOptionsDontOverrideGlobals(t *testing.T) {
	value := ""
	gfs := clitest.FlagSetterFunc(func(f *flag.FlagSet) {
		f.StringVar(&value, "value", "an actual thing", "")
	})

	sct := &SubCommanderTest{
		SubCommander: &SubCommander{
			GlobalFlags: gfs,
		},
		SubCommands: []SubCommand{
			&SubCommandStruct{
				NameValue: "sub",
			},
		},
		Args: strings.Fields("-value foobar sub"),
	}

	testSubCommanderTest(t, sct)

	if value != "foobar" {
		t.Fatal()
	}
}

func TestSubCommander_ExecuteContext_WorksCorrectlyWithGlobalOptionsAfterSubCommandAndCorrectOutputsAndCorrectErrorHappyPath(t *testing.T) {
	gfs := &clitest.SimpleFlagSetter{Suffix: "1"}
	sfs := &clitest.SimpleFlagSetter{Suffix: "2"}

	actualIn := strings.NewReader("")
	executeCalled := false
	setParametersCalled := false
	ctx := context.WithValue(context.Background(), 0, 1)
	err := fmt.Errorf("this is an error")

	sct := &SubCommanderTest{
		SubCommander: &SubCommander{
			GlobalFlags: gfs,
		},
		SubCommands: []SubCommand{
			&SubCommandStruct{
				NameValue:        "sub",
				AliasesValue:     []string{"alias"},
				SynopsisValue:    "synopsis",
				DescriptionValue: "description",
				FlagSetter:       sfs,
				ParameterSetter: &clitest.ParameterSetterStruct{
					ParameterUsageValue: func() ([]*cli.Parameter, string) {
						return []*cli.Parameter{
							&cli.Parameter{
								Name:     "p1",
								Optional: false,
								Many:     false,
							},
							&cli.Parameter{
								Name:     "p2",
								Optional: true,
								Many:     true,
							},
						}, "extra parameter usage"
					},
					SetParametersValue: func(params []string) error {
						setParametersCalled = true
						if !reflect.DeepEqual(params, []string{"foo", "bar"}) {
							t.Fatal("wrong parameters set")
						}
						return nil
					},
				},
				ExecuteValue: func(actualCtx context.Context, in io.Reader, out, outErr io.Writer) error {
					executeCalled = true
					if actualCtx != ctx {
						t.Error("did not receive correct context in execute method")
					}
					if actualIn != in {
						t.Error("did not receive correct in in execute method")
					}
					fmt.Fprintf(out, "out")
					fmt.Fprintf(outErr, "outErr")
					return err
				},
			},
		},
		Context:      ctx,
		Args:         strings.Fields("-int1 1029 -string1 whoami sub -string2 cli foo -bool2 bar"),
		In:           actualIn,
		OutString:    "out",
		OutErrString: "outErr",
		Err:          &ExecutingSubCommandError{err},
	}

	testSubCommanderTest(t, sct)

	if !executeCalled {
		t.Error("Execute() should have been called")
	}
	if !setParametersCalled {
		t.Error("SetParameters() should have been called")
	}
	if !reflect.DeepEqual(gfs, &clitest.SimpleFlagSetter{Suffix: "1", Int: 1029, String: "whoami", Bool: false}) {
		t.Error("global flags were not set correctly")
	}
	if !reflect.DeepEqual(sfs, &clitest.SimpleFlagSetter{Suffix: "2", Int: 0, String: "cli", Bool: true}) {
		t.Error("sub-command flags were not set correctly")
	}
}

func TestSubCommander_ExecuteContext_ReturnsNilErrorWhenNothingGoesWrong(t *testing.T) {
	sct := &SubCommanderTest{
		SubCommands: []SubCommand{
			&SubCommandStruct{
				NameValue:    "sub",
				ExecuteValue: clitest.NewExecuteFunc("", "", nil),
			},
		},
		Args: strings.Fields("sub"),
	}

	testSubCommanderTest(t, sct)
}

func TestSubCommander_ExecuteContext_SubCommandRegisteredHelpWillErrorParsingSubCommandParameters(t *testing.T) {
	formattedParameter := FormatParameter(&cli.Parameter{Name: SubCommandName})
	outErrStringSuffix := "\n\n" + Usage + " ... help [parameters...]" +
		"\n\n" + ParametersName + ": " + formattedParameter + "\n" +
		formattedParameter + " is the " + SubCommandName + " to provide help for" + "\n"

	tests := []struct {
		args []string
		err  error
	}{
		{
			args: strings.Fields("help"),
			err:  &cli.RequiredParameterNotSetError{Name: SubCommandName, Formatted: formattedParameter},
		},
		{
			args: strings.Fields("help sub another"),
			err:  cli.ErrTooManyParameters,
		},
	}
	for _, test := range tests {
		sct := &SubCommanderTest{
			RegisterHelp: true,
			Args:         test.args,
			OutErrString: test.err.Error() + outErrStringSuffix,
			Err:          &ParsingSubCommandError{test.err},
		}

		testSubCommanderTest(t, sct)
	}

}

func TestSubCommander_ExecuteContext_SubCommandRegisteredHelpWillErrorWithUnknownSubCommand(t *testing.T) {
	err := UnknownSubCommandError("sub")

	sct := &SubCommanderTest{
		RegisterHelp: true,
		Args:         strings.Fields("help sub"),
		OutErrString: err.Error() + "\n\n" + SimpleUsage + "\n" + SubCommandsName + ":" + "\n" +
			"  " + "help            Prints help information for a sub_command" + "\n",
		Err: &ExecutingSubCommandError{err},
	}

	testSubCommanderTest(t, sct)
}

func TestSubCommander_ExecuteContext_WorksCorrectlyWithRegisteredHelpSubCommand(t *testing.T) {
	sct := &SubCommanderTest{
		SubCommands: []SubCommand{
			&SubCommandStruct{
				NameValue: "a",
			},
			&SubCommandStruct{
				NameValue:        "sub",
				DescriptionValue: "sub_description",
			},
		},
		RegisterHelp: true,
		Args:         strings.Fields("help sub"),
		OutString:    "sub - sub_description" + "\n\n" + Usage + " ... sub" + "\n",
	}

	testSubCommanderTest(t, sct)
}

func TestSubCommander_ExecuteContext_RegisteredHelpWithHelpHelpArgsDoesNotPrintGlobalOptions(t *testing.T) {
	gfs := clitest.NewStringsFlagSetter("global1")

	sc := &SubCommander{
		GlobalFlags: gfs,
	}
	sc.RegisterHelp("help", "", "help_description")
	sct := &SubCommanderTest{
		SubCommander: sc,
		Args:         strings.Fields("help help"),
		OutString: "help - help_description\n\n" + Usage + " ... help [parameters...]\n\n" +
			ParametersName + ": " + "<SUB_COMMAND>\n" +
			"<SUB_COMMAND> is the " + SubCommandName + " to provide help for" + "\n",
	}

	testSubCommanderTest(t, sct)
}

func TestSubCommander_ExecuteContext_RegisteredHelpWithHelpListArgsDoesNotPrintGlobalOptions(t *testing.T) {
	gfs := clitest.NewStringsFlagSetter("global1")

	sc := &SubCommander{
		GlobalFlags: gfs,
	}
	sc.RegisterList("list", "", "list_description")
	sct := &SubCommanderTest{
		SubCommander: sc,
		RegisterHelp: true,
		Args:         strings.Fields("help list"),
		OutString:    "list - list_description\n\n" + Usage + " ... list\n",
	}

	testSubCommanderTest(t, sct)
}

func TestSubCommander_ExecuteContext_SubCommandRegisteredListErrorParsingSubCommandParameters(t *testing.T) {
	err := cli.ErrTooManyParameters

	sct := &SubCommanderTest{
		RegisterList: true,
		Args:         strings.Fields("list another"),
		OutErrString: err.Error() + "\n\n" + Usage + " ... list" + "\n",
		Err:          &ParsingSubCommandError{err},
	}

	testSubCommanderTest(t, sct)
}

func TestSubCommander_ExecuteContext_WorksCorrectlyWithRegsiteredListSubCommand(t *testing.T) {
	sct := &SubCommanderTest{
		RegisterList: true,
		Args:         strings.Fields("list"),
		OutString:    SubCommandsName + ":" + "\n" + "  list            Prints available sub_commands" + "\n",
	}

	testSubCommanderTest(t, sct)
}

type SubCommanderTest struct {
	*SubCommander

	SubCommands  []SubCommand
	RegisterHelp bool
	RegisterList bool

	Context context.Context
	Args    []string

	In           io.Reader
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
	prefix += "sc.ExecuteContext()"

	sc := sct.SubCommander
	if sc == nil {
		sc = &SubCommander{}
	}

	if sc.CommandName == "" {
		sc.CommandName = "command"
	}

	for _, subCommand := range sct.SubCommands {
		sc.Register(subCommand)
	}
	if sct.RegisterHelp {
		sc.RegisterHelp("help", "", "")
	}
	if sct.RegisterList {
		sc.RegisterList("list", "", "")
	}

	if sct.Context == nil {
		sct.Context = context.Background()
	}

	if sct.In == nil {
		sct.In = bytes.NewBuffer([]byte{})
	}

	out, outErr, err := executeContext(sc, sct.Context, sct.Args, sct.In)

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

func executeContext(sc *SubCommander, ctx context.Context, args []string, in io.Reader) (*bytes.Buffer, *bytes.Buffer, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	out, outErr := clitest.NewOutputs()

	err := sc.ExecuteContext(ctx, args, in, out, outErr)

	return out, outErr, err
}
