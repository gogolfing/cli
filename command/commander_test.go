package command

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

var errExecute = errors.New("error executing")

func TestSubCommander_Execute_CallsExecuteContextCorrectly(t *testing.T) {
	c := &Commander{
		Name: "command",
		Command: &CommandStruct{
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
		},
	}

	err := c.Execute(strings.Fields("a"))

	if !reflect.DeepEqual(err, &ExecutingCommandError{errExecute}) {
		t.Fatalf("err = %v WANT %v", err, errExecute)
	}
}

func TestCommander_ExecuteContext_ParsingCommandError_FlagErrHelp(t *testing.T) {
	description := "this is a description"
	prefix := "command - " + description

	ct := &CommanderTest{
		Commander: &Commander{
			Command: &CommandStruct{
				DescriptionValue: description,
			},
		},
		Args:         strings.Fields("-h"),
		OutErrString: prefix + "\n\n" + Usage + " command" + "\n",
		Err:          &ParsingCommandError{flag.ErrHelp},
	}

	testCommanderTest(t, ct)
}

func TestCommander_ExecuteContext_ParsingCommandError_OtherErrorWithOptionsOnly(t *testing.T) {
	fs := clitest.NewStringsFlagSetter("foo")
	err := errors.New("flag provided but not defined: -value")

	ct := &CommanderTest{
		Commander: &Commander{
			Command: &CommandStruct{
				DescriptionValue: "this is a description",
				FlagSetter:       fs,
			},
		},
		Args: strings.Fields("-value 12"),
		OutErrString: err.Error() + "\n\n" + Usage + " command [options...]" + "\n\n" +
			OptionsName + ":" + "\n" +
			clitest.GetFlagSetterDefaults(fs) + "\n",
		Err: &ParsingCommandError{err},
	}

	testCommanderTest(t, ct)
}

func TestCommander_ExecuteContext_ParsingCommandError_OtherErrorWithOptionsAndParametersButWithoutDescription(t *testing.T) {
	fs := clitest.NewStringsFlagSetter("foo")
	err := errors.New("flag provided but not defined: -value")

	ct := &CommanderTest{
		Commander: &Commander{
			Command: &CommandStruct{
				DescriptionValue: "this is a description",
				FlagSetter:       fs,
				ParameterSetter: &clitest.ParameterSetterStruct{
					ParameterUsageValue: func() ([]*cli.Parameter, string) {
						return []*cli.Parameter{
							&cli.Parameter{Name: "name"},
						}, "extra parameters usage"
					},
				},
			},
		},
		Args: strings.Fields("-value 12"),
		OutErrString: err.Error() + "\n\n" + Usage + " command [[options | parameters]...]" + "\n\n" +
			OptionsName + ":" + "\n" +
			clitest.GetFlagSetterDefaults(fs) + "\n\n" +
			ParametersName + ": <NAME>" + "\n" + "extra parameters usage" + "\n",
		Err: &ParsingCommandError{err},
	}

	testCommanderTest(t, ct)
}

func TestCommander_ExecuteContext_ParsingCommandError_SetParamtersErrorWithParametersButNotExtraParameterUsage(t *testing.T) {
	err := errors.New("parameter error")

	ct := &CommanderTest{
		Commander: &Commander{
			Command: &CommandStruct{
				ParameterSetter: &clitest.ParameterSetterStruct{
					SetParametersValue: func(params []string) error {
						return err
					},
					ParameterUsageValue: func() ([]*cli.Parameter, string) {
						return []*cli.Parameter{
							&cli.Parameter{Name: "name"},
						}, ""
					},
				},
			},
		},
		Args: strings.Fields("foobar"),
		OutErrString: err.Error() + "\n\n" + Usage + " command [parameters...]" + "\n\n" +
			ParametersName + ": <NAME>" + "\n",
		Err: &ParsingCommandError{err},
	}

	testCommanderTest(t, ct)
}

func TestCommander_ExecuteContext_ParsingCommandError_StillPrintsExtraUsageOnNoParameterSetError(t *testing.T) {
	err := cli.ErrTooManyParameters

	ct := &CommanderTest{
		Commander: &Commander{
			Command: &CommandStruct{
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
		Args: strings.Fields("foobar"),
		OutErrString: err.Error() + "\n\n" + Usage + " command" + "\n\n" +
			"extra parameter usage" + "\n",
		Err: &ParsingCommandError{err},
	}

	testCommanderTest(t, ct)
}

func TestCommander_ExecuteContext_ExecutionError(t *testing.T) {
	err := errExecute

	ct := &CommanderTest{
		Commander: &Commander{
			Command: &CommandStruct{
				ExecuteValue: clitest.NewExecuteFunc("", "", err),
			},
		},
		Err: &ExecutingCommandError{err},
	}

	testCommanderTest(t, ct)
}

func TestCommander_ExecuteContext_WorksCorrectly(t *testing.T) {
	err := errExecute
	fs := &clitest.SimpleFlagSetter{}

	actualIn := strings.NewReader("in")
	setParametersCalled, executeCalled := false, false
	ctx := context.WithValue(context.Background(), 0, 1)

	ct := &CommanderTest{
		Commander: &Commander{
			Command: &CommandStruct{
				FlagSetter: fs,
				ParameterSetter: &clitest.ParameterSetterStruct{
					SetParametersValue: func(params []string) error {
						setParametersCalled = true
						if !reflect.DeepEqual(params, []string{"foo", "bar", "hello", "-world"}) {
							t.Fatal("incorrect parameters")
						}
						return nil
					},
				},
				ExecuteValue: func(actualCtx context.Context, in io.Reader, out io.Writer, outErr io.Writer) error {
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
		Args:         strings.Fields("foo -int 1234 -string value -bool bar -- hello -world"),
		In:           actualIn,
		OutString:    "out",
		OutErrString: "outErr",
		Err:          &ExecutingCommandError{err},
	}

	testCommanderTest(t, ct)

	if !setParametersCalled {
		t.Error("SetParameters() should have been called")
	}
	if !executeCalled {
		t.Error("Execute() should have been called")
	}
	if !reflect.DeepEqual(fs, &clitest.SimpleFlagSetter{Int: 1234, String: "value", Bool: true}) {
		t.Error("flags were not set correctly")
	}
}

func TestCommander_ExecuteContext_ReturnsNilErrorWhenNothingGoesWrong(t *testing.T) {
	ct := &CommanderTest{
		Commander: &Commander{
			Command: &CommandStruct{
				ExecuteValue: clitest.NewExecuteFunc("", "", nil),
			},
		},
	}

	testCommanderTest(t, ct)
}

type CommanderTest struct {
	*Commander

	Context context.Context
	Args    []string

	In io.Reader

	OutString    string
	OutErrString string
	Err          error
}

func testCommanderTest(t *testing.T, ct *CommanderTest) {
	prefix := fmt.Sprintf("%s: %s", t.Name(), "c.ExecuteContext()")

	if ct.Commander.Name == "" {
		ct.Commander.Name = "command"
	}

	if ct.Context == nil {
		ct.Context = context.Background()
	}

	if ct.In == nil {
		ct.In = bytes.NewBuffer([]byte{})
	}

	out, outErr, err := executeContext(ct.Commander, ct.Context, ct.Args, ct.In)

	outString := out.String()
	outErrString := outErr.String()

	if outString != ct.OutString {
		t.Errorf(
			"%v: out = %v WANT %v",
			t.Name(),
			outString,
			ct.OutString,
		)
	}

	if outErrString != ct.OutErrString {
		t.Errorf(
			"%v: outErr = %v WANT %v",
			prefix,
			outErrString,
			ct.OutErrString,
		)
	}

	if !reflect.DeepEqual(err, ct.Err) {
		t.Errorf(
			"%v: err = %v WANT %v",
			prefix,
			err,
			ct.Err,
		)
	}
}

func executeContext(c *Commander, ctx context.Context, args []string, in io.Reader) (*bytes.Buffer, *bytes.Buffer, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	out, outErr := clitest.NewOutputs()

	err := c.ExecuteContext(ctx, args, in, out, outErr)

	return out, outErr, err
}
