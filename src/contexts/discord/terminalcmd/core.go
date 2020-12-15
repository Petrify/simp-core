package terminalcmd

import (
	"errors"
)

type action func(*Context) (c *Context, err error)

var actions map[string]action = make(map[string]action)

func registerAction(name string, act action) (err error) {
	actions[name] = act
	return
}

func (ctx *Context) DoAction(action string, args []interface{}) (c *Context, err error) {
	act, ok := actions[action]
	if !ok {
		err = errors.New("Unsupported Action: " + action)
		c = ctx
		return
	}
	c, err = act(ctx)
	return
}
