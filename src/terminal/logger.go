package terminal

import (
	"fmt"
)

type Logger interface {
	Log(a ...interface{})
	Warn(a ...interface{})
	Err(a ...interface{})
}

type StdLogger struct {
	term  Terminal
	label string
}

func newStdLogger(t Terminal, label string) *StdLogger {
	return &StdLogger{t, label}
}

func (l *StdLogger) Log(a ...interface{}) {
	s := fmt.Sprintf("[%s] %s", l.label, fmt.Sprint(a...))
	l.term.Print(s)
}

func (l *StdLogger) Logf(pattern string, a ...interface{}) {
	s := fmt.Sprintf(pattern, a...)
	l.term.Print(s)
}

func (l *StdLogger) Warn(a ...interface{}) { //TODO: Add Color
	s := fmt.Sprintf("[%s] Warning: %s", l.label, fmt.Sprint(a...))
	l.term.Print(s)
}

func (l *StdLogger) Err(a ...interface{}) { //TODO: Add color
	s := fmt.Sprintf("[%s] Error: %s", l.label, fmt.Sprint(a...))
	l.term.Print(s)
}
