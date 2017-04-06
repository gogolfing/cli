package command

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gogolfing/cli"
)

//Commander provides options and methods for executing a Command.
//
//Note that Commander is NOT safe for use with multiple goroutines.
type Commander struct {
	//The program name to execute. Used in help and error output.
	//This will usually be os.Args[0].
	Name string

	//Command is the Command to execute.
	Command
}

//Execute is syntactic sugar for ExecuteContext() with context.Background(), args,
//os.Stdin, os.Stdout, and os.Stderr.
func (c *Commander) Execute(args []string) error {
	return c.ExecuteContext(context.Background(), args, os.Stdin, os.Stdout, os.Stderr)
}

//ExecuteContext executes c.Command with the provided parameters.
//
//Ctx is the Context passed unaltered to c.Command.Execute.
//
//Args should be the program arguments excluding the program name - usually os.Args[1:].
//They will be parsed using cli.ParseArgumentsInterspersed.
//
//The parameters in, out, and outErr are passed unaltered to c.Command.Execute
//and should represent the standard input, output, and error files for the executing
//Command.
//
//Err will be non-nil if parsing args failed - with type *ParsingCommandError
//or *ExecutingCommandError if the c.Command.Execute method returns an error.
//
//If the returned error is of type *ParsingCommandError, then error and help output
//will be written to outErr. See the package documentation for more details on error
//and help output. If this is the error, then execution stops and c.Command.Execute
//is never called.
//
//If the error is an *ExecutingCommandError then nothing is output by c.
func (c *Commander) ExecuteContext(ctx context.Context, args []string, in io.Reader, out, outErr io.Writer) error {
	err := c.executeContext(ctx, args, in, out, outErr)
	if err == nil {
		return nil
	}

	if pce, ok := err.(*ParsingCommandError); ok {
		if pce.Err == flag.ErrHelp {
			c.printCommandError(outErr, nil, true)
		} else {
			c.printCommandError(outErr, pce, false)
		}
	}

	return err
}

func (c *Commander) executeContext(ctx context.Context, args []string, in io.Reader, out, outErr io.Writer) error {
	f := cli.NewFlagSet(c.Name, c)

	params, err := cli.ParseArgumentsInterspersed(f, args)
	if err != nil {
		return &ParsingCommandError{err}
	}
	if err := c.SetParameters(params); err != nil {
		return &ParsingCommandError{err}
	}

	if err := c.Command.Execute(ctx, in, out, outErr); err != nil {
		return &ExecutingCommandError{err}
	}

	return nil
}

func (c *Commander) printCommandError(out io.Writer, err error, description bool) {
	if err != nil {
		fmt.Fprintf(out, "%v\n\n", err)
	}

	if description {
		c.maybePrintHeader(out)
	}

	c.printCommandUsage(out)
}

func (c *Commander) maybePrintHeader(out io.Writer) {
	if description := c.Description(); len(description) > 0 {
		fmt.Fprintf(out, "%s - %s\n\n", c.Name, description)
	}
}

func (c *Commander) printCommandUsage(out io.Writer) {
	fmt.Fprintf(out, "%s %s", Usage, c.Name)

	c.maybePrintCommandLineUsage(out)

	fmt.Fprintln(out)

	c.maybePrintOptionsUsage(out)
	c.maybePrintParameterUsage(out)
}

func (c *Commander) maybePrintCommandLineUsage(out io.Writer) {
	if commandLineUsage := c.getCommandLineUsage(); len(commandLineUsage) > 0 {
		fmt.Fprintf(out, " %s", commandLineUsage)
	}
}

func (c *Commander) getCommandLineUsage() string {
	hasOptions := cli.CountFlags(cli.NewFlagSet(c.Name, c)) > 0
	params, _ := c.ParameterUsage()
	hasParameters := len(params) > 0

	args := []string{}
	if hasOptions {
		args = append(args, OptionsName)
	}
	if hasParameters {
		args = append(args, ParametersName)
	}

	if len(args) == 0 {
		return ""
	}
	joined := strings.Join(args, ArgumentSeparator)
	if len(args) == 1 {
		return FormatArgument(joined, true, true)
	}
	return FormatArgument(FormatArgument(joined, true, false), true, true)
}

func (c *Commander) maybePrintOptionsUsage(out io.Writer) {
	if !c.hasOptions() {
		return
	}
	fmt.Fprintf(out, "\n%s:\n%s\n", OptionsName, cli.GetFlagSetDefaults(cli.NewFlagSet(c.Name, c)))
}

func (c *Commander) maybePrintParameterUsage(out io.Writer) {
	params, usage := c.ParameterUsage()

	didPrint := false
	if formatted := cli.FormatParameters(params, FormatParameter); len(formatted) > 0 {
		fmt.Fprintf(out, "\n%s: %s", ParametersName, formatted)
		didPrint = true
	}
	if len(usage) > 0 {
		fmt.Fprintf(out, "\n%s", usage)
		didPrint = true
	}
	if didPrint {
		fmt.Fprintln(out)
	}
}

func (c *Commander) hasOptions() bool {
	return cli.CountFlags(cli.NewFlagSet(c.Name, c)) > 0
}
