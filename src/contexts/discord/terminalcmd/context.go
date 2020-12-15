package terminalcmd

type Context struct {
	command string
}

func (ctx *Context) Command() (cmd string) {
	return ctx.command //TODO:
}
