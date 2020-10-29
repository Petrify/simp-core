package discord

import (
	"flag"
	"fmt"
	"schoolbot/internal/model"

	"github.com/objectbox/objectbox-go/objectbox"
)

type catalog []*model.Class

func (c catalog) String(i int) string {
	return fmt.Sprintf("%s %s", c[i].Abbr, c[i].Name)
}

func (c catalog) Len() int {
	return len((c))
}

// Variables used for command line parameters
var (
	Token string
)

// Variables for bot state
var (
	GCManager *GuildManager
	TManager  *TerminalManager
)

func init() {

	flag.StringVar(&Token, "t", "", "Bot Token")

}

func initObjectBox() *objectbox.ObjectBox {
	objectBox, err := objectbox.NewBuilder().Model(model.ObjectBoxModel()).Build()
	if err != nil {
		panic(err)
	}
	return objectBox
}

func Run() {

	TManager, _ = NewTerminalManager(Token)
	GCManager, _ = NewGuildManager(Token, initObjectBox())

}
