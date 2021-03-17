package commands

import (
	"fmt"
	"strings"
)

type InvalidCommandError struct {
	cmd   string
	depth int
}

func (e InvalidCommandError) Error() string {
	return fmt.Sprintf("Unknown Command: %s", e.cmd)
}

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
