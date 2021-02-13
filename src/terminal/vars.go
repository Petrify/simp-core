package terminal

import "os"

var SysTerminal Terminal

func init() {
	SysTerminal = NewStdTerminal(os.Stdin, os.Stdout)
}
