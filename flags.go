package main

import "flag"

var (
	token string
)

func initFlags() {
	flag.StringVar(&token, "t", "", "Bot Token")
}
