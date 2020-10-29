package discord

import (
	"testing"
	"time"
)

var (
	gm *GuildManager
	tm *TerminalManager
)

func TestGC(T *testing.T) {
	gm, _ = NewGuildManager("Njk1NTk1MDUwMjQyOTMyODY4.XocdXw.Tynx3R1aJ6WpOW3b_m4o0dLfbq0", initObjectBox())
	tm, _ = NewTerminalManager("Njk1NTk1MDUwMjQyOTMyODY4.XocdXw.Tynx3R1aJ6WpOW3b_m4o0dLfbq0")

	dura, _ := time.ParseDuration("180s")
	t := time.NewTimer(dura)
	<-t.C

	tm.Close()
	gm.Close()
}
