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
	//This will usuaully be os.Args[0].
	Name string

	//Command is the Command to execute.
	Command
}

//Execute is syntactic sugar for ExecuteContext() with context.Background().
func (c *Commander) Execute(args []string) error {
	return c.ExecuteContext(context.Background(), args)
}

//ExecuteContext is syntactic sugar for ExecuteContextOut() with os.Stdout and os.Stderr.
func (c *Commander) ExecuteContext(ctx context.Context, args []string) error {
	return c.ExecuteContextOut(ctx, args, os.Stdout, os.Stderr)
}

//ExecuteContextOut attempts to execute c.Command.
//ctx will be passed along unaltered to the Command's Execute() method.
//args are the command line arguments to parse and use for Command execution.
//They should include command line arguments excluding the program name - usually os.Args[1:].
//out and outErr are the io.Writers to use for standard out and standard error
//for Command execution and help and error output.
//
//TODO
func (c *Commander) ExecuteContextOut(ctx context.Context, args []string, out, outErr io.Writer) error {
	err := c.executeContextOut(ctx, args, out, outErr)
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

func (c *Commander) executeContextOut(ctx context.Context, args []string, out, outErr io.Writer) error {
	f := cli.NewFlagSet(c.Name, c)

	params, err := cli.ParseArgumentsInterspersed(f, args)
	if err != nil {
		return &ParsingCommandError{err}
	}
	if err := c.SetParameters(params); err != nil {
		return &ParsingCommandError{err}
	}

	if err := c.Command.Execute(ctx, out, outErr); err != nil {
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
	if len(params) == 0 {
		return
	}

	fmt.Fprintf(out, "\n%s: %s\n", ParametersName, cli.FormatParameters(params, FormatParameter))
	if len(usage) > 0 {
		fmt.Fprintf(out, "%s\n", usage)
	}
}

func (c *Commander) hasOptions() bool {
	return cli.CountFlags(cli.NewFlagSet(c.Name, c)) > 0
}
