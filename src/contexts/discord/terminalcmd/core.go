package terminalcmd

import (
	"errors"
)

type action struct {
	func(*Context, ...interface{}) (c *Context, err error)

var actions map[string]action = make(map[string]action)

func registerAction(name string, act action) (err error) {
	actions[name] = act
	return
}

func (ctx *Context) ExecScript(script string, args []interface{}) (c *Context, err error) {
	act, ok := actions[action]
	if !ok {
		err = errors.New("Unsupported Action: " + action)
		c = ctx
		return
	}
	c, err = act(ctx)
	return
}

func tokenizeScript(script string) (root *token) {
	root = &token{
		"",
		make([]interface{}, 10)
	}

	

	return
}

type token struct {
	key string
	args []interface{}
}

func (t *token) validate()
