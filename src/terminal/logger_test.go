package terminal

import "testing"

func Test_logger(T *testing.T) {
	term := SysTerminal
	log := term.Logger("Test Label")
	log.Log("Hello, World")
	log.Warn("Warn Test")
	log.Err("Err Test")

}
