package commands

import (
	"context"
	"errors"
	"strings"
)

//Command interpreter for guilds
type Interpreter struct {
	root *commandNode
}

func NewInterpreter() *Interpreter {
	return &Interpreter{
		root: &commandNode{
			"", //root node has no key
			make(map[string]*commandNode),
			nil,
		},
	}
}

func (it *Interpreter) Run(ctx context.Context, cmd string) (err error) {

	splitCmd := strings.Split(cmd, " ")
	curr := it.root

	var (
		next    *commandNode
		exisits bool
		depth   int
	)

	for depth, key := range splitCmd {

		if curr.children == nil {
			err = errors.New(("Command Interpreter error: Invalid command node"))
			return
		} else {
			next, exisits = curr.children[key]
		}

		if !exisits {
			err = InvalidCommandError{cmd, depth}
			return
		}

		curr = next

		if curr.f != nil {
			break
		}
	}

	curr.f(ctx, splitCmd[depth:])

	return
}

func (it *Interpreter) AddCommand(path string, f func(ctx context.Context, args []string)) (err error) {
	splitCmd := strings.Split(path, " ")
	curr := it.root
	var (
		next    *commandNode
		exisits bool
	)
	for _, key := range splitCmd {
		if curr.f != nil {
			return errors.New("Can not create command as subcommand of an existing command")
		}

		if curr.children == nil {
			exisits = false
			curr.children = make(map[string]*commandNode)
		} else {
			next, exisits = curr.children[key]
		}

		//if child does not yet exist, create one
		if !exisits {
			next = &commandNode{
				key,
				nil,
				nil,
			}
			curr.children[key] = next
		}
		curr = next
	}

	//if we got this far, the command is valid and the last commandNode is stored in curr
	curr.f = f
	return
}

//-----Command Tree structure------

type commandNode struct {
	key      string
	children map[string]*commandNode
	f        func(ctx context.Context, args []string)
}
