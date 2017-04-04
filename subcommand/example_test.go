package subcommand

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gogolfing/cli"
	"github.com/gogolfing/cli/clitest"
)

func Example() {
	subCommand := &SubCommandStruct{
		NameValue: "sub1",
		ExecuteValue: func(ctx context.Context, _ io.Reader, out, _ io.Writer) error {
			fs := ctx.Value(0).(*clitest.SimpleFlagSetter)
			fmt.Fprintln(out, fs.Int)
			return nil
		},
	}

	globalFlags := &clitest.SimpleFlagSetter{}

	sc := &SubCommander{
		GlobalFlags: globalFlags,
	}
	sc.Register(subCommand)

	ctx := context.WithValue(context.Background(), 0, globalFlags)

	err := sc.ExecuteContext(
		ctx,
		strings.Fields("-int 1234 sub1"),
		os.Stdin,
		os.Stdout,
		os.Stderr,
	)
	if err != nil {
		fmt.Println(err)
	}

	// Output:
	// 1234
}

func Example_errorFlagErrHelp() {
	subCommand1 := &SubCommandStruct{
		NameValue:     "sub1",
		SynopsisValue: "Synopsis for sub1",
	}
	subCommand2 := &SubCommandStruct{
		NameValue:     "sub2",
		AliasesValue:  []string{"alias2.2", "alias2.1"},
		SynopsisValue: "Synopsis for sub2",
	}

	sc := &SubCommander{
		CommandName: "example_errorHelp",
		GlobalFlags: clitest.NewStringsFlagSetter("value"),
	}
	sc.Register(subCommand2)
	sc.Register(subCommand1)

	//We use os.Stdout for outErr so that we can verify output.
	sc.ExecuteContext(
		context.Background(),
		strings.Fields("-h"),
		os.Stdin,
		os.Stdout,
		os.Stdout,
	)

	// Output:
	// usage: example_errorHelp [global_options...] <sub_command> [[global_options | sub_command_options | parameters]...]
	//
	// global_options:
	//   -value string
	//     	value_usage (default "value_default")
	//
	// sub_commands:
	//   sub1                        Synopsis for sub1
	//   sub2, alias2.1, alias2.2    Synopsis for sub2
}

func Example_errorParsingGlobalArguments() {
	subCommand1 := &SubCommandStruct{
		NameValue:     "sub1",
		SynopsisValue: "Synopsis for sub1",
	}

	sc := &SubCommander{
		CommandName: "example_errorParsingGlobalArguments",
		GlobalFlags: clitest.NewStringsFlagSetter("value"),
	}
	sc.Register(subCommand1)

	//We use os.Stdout for outErr so that we can verify output.
	sc.ExecuteContext(
		context.Background(),
		strings.Fields("-foo bar sub1"),
		os.Stdin,
		os.Stdout,
		os.Stdout,
	)

	// Output:
	// flag provided but not defined: -foo
	//
	// usage: example_errorParsingGlobalArguments [global_options...] <sub_command> [[global_options | sub_command_options | parameters]...]
	//
	// global_options:
	//   -value string
	//     	value_usage (default "value_default")
	//
	// sub_commands:
	//   sub1            Synopsis for sub1
}

func Example_errorUnknownSubCommand() {
	subCommand1 := &SubCommandStruct{
		NameValue:     "sub1",
		SynopsisValue: "Synopsis for sub1",
	}

	sc := &SubCommander{
		CommandName: "example_errorUnknownSubCommand",
		GlobalFlags: clitest.NewStringsFlagSetter("value"),
	}
	sc.Register(subCommand1)

	//We use os.Stdout for outErr so that we can verify output.
	sc.ExecuteContext(
		context.Background(),
		strings.Fields("-value foobar sub2"),
		os.Stdin,
		os.Stdout,
		os.Stdout,
	)

	// Output:
	// unknown sub_command "sub2"
	//
	// usage: example_errorUnknownSubCommand [global_options...] <sub_command> [[global_options | sub_command_options | parameters]...]
	//
	// sub_commands:
	//   sub1            Synopsis for sub1
}

func Example_errorParsingSubCommandArguments() {
	subCommand1 := &SubCommandStruct{
		NameValue:     "sub1",
		SynopsisValue: "Synopsis for sub1",
		FlagSetter:    clitest.NewStringsFlagSetter("subflag"),
	}

	sc := &SubCommander{
		CommandName: "example_errorParsingSubCommandArguments",
		GlobalFlags: clitest.NewStringsFlagSetter("value"),
	}
	sc.Register(subCommand1)

	//We use os.Stdout for outErr so that we can verify output.
	sc.ExecuteContext(
		context.Background(),
		strings.Fields("-value foobar sub1 -foo bar"),
		os.Stdin,
		os.Stdout,
		os.Stdout,
	)

	// Output:
	// flag provided but not defined: -foo
	//
	// usage: ... sub1 [[global_options | sub_command_options]...]
	//
	// global_options:
	//   -value string
	//     	value_usage (default "value_default")
	//
	// sub_command_options:
	//   -subflag string
	//     	subflag_usage (default "subflag_default")
}

func Example_helpSubCommand() {
	subCommand1 := &SubCommandStruct{
		NameValue:        "sub1",
		SynopsisValue:    "Synopsis for sub1",
		DescriptionValue: "This is a description.",
		ParameterSetter: &clitest.ParameterSetterStruct{
			ParameterUsageValue: func() ([]*cli.Parameter, string) {
				params := []*cli.Parameter{
					&cli.Parameter{
						Name: "files",
						Many: true,
					},
				}
				return params, FormatParameter(params[0]) + " are the files to process"
			},
		},
	}

	sc := &SubCommander{
		CommandName: "example_helpSubCommand",
	}
	sc.RegisterHelp("help", "", "")
	sc.Register(subCommand1)

	//We use os.Stdout for outErr so that we can verify output.
	sc.ExecuteContext(
		context.Background(),
		strings.Fields("help sub1"),
		os.Stdin,
		os.Stdout,
		os.Stdout,
	)

	// Output:
	// sub1 - This is a description.
	// usage: ... sub1 [parameters...]

	// parameters: <FILES...>
	// <FILES...> are the files to process
}

func Example_listSubCommand() {
	subCommand1 := &SubCommandStruct{
		NameValue:     "sub1",
		SynopsisValue: "Synopsis for sub1",
	}

	sc := &SubCommander{
		CommandName: "example_listSubCommand",
	}
	sc.RegisterHelp("help", "", "", "h")
	sc.RegisterList("list", "", "", "subcommands")
	sc.Register(subCommand1)

	//We use os.Stdout for outErr so that we can verify output.
	sc.ExecuteContext(
		context.Background(),
		strings.Fields("list"),
		os.Stdin,
		os.Stdout,
		os.Stdout,
	)

	// Output:
	// sub_commands:
	//   help, h              Prints help information for a sub_command
	//   list, subcommands    Prints available sub_commands
	//   sub1                 Synopsis for sub1
}
