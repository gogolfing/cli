package command

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

	"github.com/gogolfing/cli"
	"github.com/gogolfing/cli/clitest"
)

var errExecute = errors.New("error executing")

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

	c := &Commander{
		Name: "command",
		Command: &CommandStruct{
			ExecuteValue: clitest.NewExecuteFunc("out", "outErr", errExecute),
		},
	}

	err := c.Execute(strings.Fields("a"))

	out.Close()
	outErr.Close()

	if outBytes, _ := ioutil.ReadFile(os.Stdout.Name()); string(outBytes) != "out" {
		t.Fatalf("out = %v WANT %s", string(outBytes), "out")
	}
	if outErrBytes, _ := ioutil.ReadFile(os.Stderr.Name()); string(outErrBytes) != "outErr" {
		t.Fatalf("outErr = %s WANT %s", outErrBytes, "outErr")
	}
	if !reflect.DeepEqual(err, &ExecutingCommandError{errExecute}) {
		t.Fatalf("err = %v WANT %v", err, errExecute)
	}
}

func TestCommander_ExecuteContextOut_ParsingCommandError_FlagErrHelp(t *testing.T) {
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

func TestCommander_ExecuteContextOut_ParsingCommandError_OtherErrorWithOptionsOnly(t *testing.T) {
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

func TestCommander_ExecuteContextOut_ParsingCommandError_OtherErrorWithOptionsAndParametersButWithoutDescription(t *testing.T) {
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

func TestCommander_ExecuteContextOut_ParsingCommandError_SetParamtersErrorWithParametersButNotExtraParameterUsage(t *testing.T) {
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

func TestCommander_ExecuteContextOut_ExecutionError(t *testing.T) {
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

func TestCommander_ExecuteContextOut_WorksCorrectly(t *testing.T) {
	err := errExecute
	fs := &clitest.SimpleFlagSetter{}

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
							t.Fatal("incorrect paramters")
						}
						return nil
					},
				},
				ExecuteValue: func(actualCtx context.Context, out io.Writer, outErr io.Writer) error {
					executeCalled = true
					if actualCtx != ctx {
						t.Error("did not receive correct context in execute method")
					}
					fmt.Fprintf(out, "out")
					fmt.Fprintf(outErr, "outErr")
					return err
				},
			},
		},
		Context:      ctx,
		Args:         strings.Fields("foo -int 1234 -string value -bool bar -- hello -world"),
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

type CommanderTest struct {
	*Commander

	Context context.Context
	Args    []string

	OutString    string
	OutErrString string
	Err          error
}

func testCommanderTest(t *testing.T, ct *CommanderTest) {
	prefix := fmt.Sprintf("%s: %s", t.Name(), "c.ExecuteContextOut()")

	if ct.Commander.Name == "" {
		ct.Commander.Name = "command"
	}

	if ct.Context == nil {
		ct.Context = context.Background()
	}

	out, outErr, err := executeContextOut(ct.Commander, ct.Context, ct.Args)

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

func executeContextOut(c *Commander, ctx context.Context, args []string) (*bytes.Buffer, *bytes.Buffer, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	out, outErr := clitest.NewOutputs()

	err := c.ExecuteContextOut(ctx, args, out, outErr)

	return out, outErr, err
}
