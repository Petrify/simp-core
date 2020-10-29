package discord

import "strings"

//Error interface thrown by Commands when something goes wrong
type CommandError interface {
	Error() string
	setPath(path []string)
}

//Implements Command Error
type InvalidCommandError struct {
	Path []string
}

func (e *InvalidCommandError) Error() string {
	builder := strings.Builder{}
	builder.WriteString("Unknown Command: ")
	builder.WriteString(strings.Join(e.Path, " "))
	return builder.String()
}

func (e *InvalidCommandError) setPath(path []string) {
	e.Path = path
}

//Implements Command Error
type InvalidArgsError struct {
	Path []string
}

func (e *InvalidArgsError) Error() string {
	builder := strings.Builder{}
	builder.WriteString("Invalid arguments for Command: ")
	builder.WriteString(strings.Join(e.Path, " "))
	return builder.String()
}

func (e *InvalidArgsError) setPath(path []string) {
	e.Path = path

}

//Implements Command Error
type ExecutionError struct {
	Path   []string
	CmdErr error
}

func (e *ExecutionError) Error() string {
	builder := strings.Builder{}
	builder.WriteString("Error during execution of \"")
	builder.WriteString(strings.Join(e.Path, " "))
	builder.WriteString("\" Encountered: ")
	builder.WriteString(e.CmdErr.Error())
	return builder.String()
}

func (e *ExecutionError) setPath(path []string) {
	e.Path = path
}
