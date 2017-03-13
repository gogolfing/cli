package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
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
		t.Fatalf("out = %v WANT %s", outBytes, "out")
	}
	if outErrBytes, _ := ioutil.ReadFile(os.Stderr.Name()); string(outErrBytes) != "outErr" {
		t.Fatalf("outErr = %s WANT %s", outErrBytes, "outErr")
	}
	if err.Error() != errExecute.Error() {
		t.Fatalf("err = %v WANT %v", err, errExecute)
	}
}

func TestSubCommander_ExecuteContextOut_ErrUnsuppliedSubCommmand(t *testing.T) {
	outErrString := "sub_command not supplied\n\n"

	tests := []*SubCommanderTest{
		{
			Args:         nil,
			OutErrString: outErrString,
			Err:          ErrUnsuppliedSubCommand,
		},
		{
			Args:         []string{},
			OutErrString: outErrString,
			Err:          ErrUnsuppliedSubCommand,
		},
		{
			SubCommander: &SubCommander{
				GlobalFlags: NewStringsFlagSetter("value"),
			},
			Args:         []string{"-value", "1234"},
			OutErrString: outErrString,
			Err:          ErrUnsuppliedSubCommand,
		},
	}

	testSubCommanderTests(t, tests)
}

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
