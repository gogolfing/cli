package command

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gogolfing/cli"
	"github.com/gogolfing/cli/clitest"
)

func Example() {
	count := 0
	fs := clitest.FlagSetterFunc(func(f *flag.FlagSet) {
		f.IntVar(&count, "count", count, "number of times to print parameters")
	})

	var params []string = nil
	command := &CommandStruct{
		DescriptionValue: "example description",
		FlagSetter:       fs,
		ParameterSetter: &clitest.ParameterSetterStruct{
			SetParametersValue: func(p []string) error {
				params = append([]string{}, p...)
				return nil
			},
		},
		ExecuteValue: func(_ context.Context, _ io.Reader, out, _ io.Writer) error {
			for _, p := range params {
				for c := 0; c < count; c++ {
					fmt.Fprintf(out, "%s", p)
				}
			}
			fmt.Fprintln(out)
			return nil
		},
	}

	commander := &Commander{
		Name:    "example",
		Command: command,
	}

	err := commander.Execute(strings.Fields("a -count 2 b"))
	if err != nil {
		fmt.Println(err)
	}

	//Output:
	//aabb
}

func ExampleErrorFlagErrHelp() {
	command := &CommandStruct{
		DescriptionValue: "this is a description.",
		ParameterSetter: &clitest.ParameterSetterStruct{
			ParameterUsageValue: func() ([]*cli.Parameter, string) {
				return []*cli.Parameter{
					&cli.Parameter{
						Name: "p1",
					},
					&cli.Parameter{
						Name:     "p2s",
						Optional: true,
						Many:     true,
					},
				}, "extra parameter usage"
			},
		},
	}

	commander := &Commander{
		Name:    "example_help",
		Command: command,
	}

	//We set the outErr paramter to os.Stdout so that we can validate the test output.
	commander.ExecuteContext(
		context.Background(),
		strings.Fields("-h"),
		os.Stdin,
		os.Stdout,
		os.Stdout,
	)

	// Output:
	// example_help - this is a description.
	//
	// usage: example_help [parameters...]
	//
	// parameters: <P1> [P2S...]
	// extra parameter usage
}

func ExampleErrorParsingArguments() {
	fs := clitest.FlagSetterFunc(func(f *flag.FlagSet) {
		f.Int("count", 0, "number of times to print parameters")
	})

	command := &CommandStruct{
		DescriptionValue: "this is a description.",
		FlagSetter:       fs,

		//Note that we do not set ExecuteValue. It will not be called since there
		//is an argument parsing error.
	}

	commander := &Commander{
		Name:    "example_error",
		Command: command,
	}

	//We set the outErr paramter to os.Stdout so that we can validate the test output.
	commander.ExecuteContext(
		context.Background(),
		strings.Fields("-value foobar"),
		os.Stdin,
		os.Stdout,
		os.Stdout,
	)

	//The Output: below may look weird because of the tab and space formatting
	//and what is required by the testing package.

	// Output:
	// flag provided but not defined: -value
	//
	// usage: example_error [options...]
	//
	// options:
	//   -count int
	//     	number of times to print parameters
}
