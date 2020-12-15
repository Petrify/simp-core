package main

import "flag"

var (
	token string
)

func init() {

	flag.StringVar(&Token, "t", "", "Bot Token")

}
