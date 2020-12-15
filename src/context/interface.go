package context

type Context interface {
	DoAction(act string, args []interface{}) (c *Context, err error)
}
