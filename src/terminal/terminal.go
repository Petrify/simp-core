package terminal

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"
)

var ()

/* A terminal should:
+ have print functions
have channel for queued inputs
+ can create a logger (single direction logging with source labels)
+ logs time of printing


*/
type Terminal interface {
	Print(s ...interface{}) error
	Printf(format string, a ...interface{}) error
	Readln() (string, error)
	Logger(label string) Logger
}

type stdTerminal struct {
	scanner   *bufio.Scanner
	writeLock *sync.Mutex
}

func NewStdTerminal(in *os.File, out *os.File) *stdTerminal {
	return &stdTerminal{
		bufio.NewScanner(in),
		&sync.Mutex{},
	}
}

func (t *stdTerminal) Print(a ...interface{}) error {

	t.writeLock.Lock()
	defer t.writeLock.Unlock()

	s := format(fmt.Sprint(a...))
	_, e := os.Stdout.WriteString(s)
	return e
}

func (t *stdTerminal) Printf(format string, a ...interface{}) error {
	s := fmt.Sprintf(format, a...)
	return t.Print(s)
}

func (t *stdTerminal) Logger(label string) Logger {
	return newStdLogger(t, label)
}

func (t *stdTerminal) Readln() (s string, e error) {
	if !t.scanner.Scan() {
		return "", t.scanner.Err()
	}

	s = t.scanner.Text()
	return s, t.scanner.Err()
}

func format(a string) string {
	return fmt.Sprintf("[%s] %s\n", time.Now().Format("15:04:05"), a)
}
