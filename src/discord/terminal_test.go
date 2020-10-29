package discord

import (
	"errors"
	"testing"
	"time"
)

func TestTerminal(T *testing.T) {
	m, err := NewTerminalManager("Njk1NTk1MDUwMjQyOTMyODY4.XocdXw.Tynx3R1aJ6WpOW3b_m4o0dLfbq0")
	if err != nil {
	}

	I := NewTerminalInterpreter()
	I.AddCommand("echo", echo)
	I.AddCommand("test", test)

	m.NewTerminal("84787975480700928", I, "30s", "Test terminal active")
	dura, _ := time.ParseDuration("180s")
	t := time.NewTimer(dura)
	<-t.C
	m.Close()
}

func echo(inp []string, t *TerminalSession) error {
	t.Write("What to echo?")
	echo, ok := t.Read()
	if !ok {
		return errors.New("Test Error")
	}
	return t.Write(echo)
}

func test(inp []string, t *TerminalSession) error {
	return errors.New("Test Error")
}
