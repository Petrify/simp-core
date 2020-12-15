package commands

import (
	"fmt"
	"strings"
)

//Error interface thrown by Commands when something goes wrong
type CommandError interface {
	error
}

//Implements Command Error
type InvalidCommandError struct {
	cmd   string
	depth int
}

func (e InvalidCommandError) Error() string {
	return fmt.Sprintf("Unknown Command: %s", e.cmd)
}

//Implements Command Error
type InvalidArgsError struct {
	cmd   string
	depth int
}

func (e InvalidArgsError) Error() string {

	return fmt.Sprintf("Invalid arguments for Command: %s", e.cmd)

}

func (e InvalidArgsError) Depth() int {
	return e.depth
}

//Implements Command Error
type ExecutionError struct {
	cmd    string
	depth  int
	CmdErr error
}

func (e ExecutionError) Error() string {
	builder := strings.Builder{}
	builder.WriteString("Error during execution of \"")
	builder.WriteString(e.cmd)
	builder.WriteString("\" Encountered: ")
	builder.WriteString(e.CmdErr.Error())
	return builder.String()
}
