package commands

import (
	"errors"
	"fmt"
	"strings"
)

//Command interpreter for guilds
type Interpreter struct {
	root *commandNode
}

func NewInterpreter() *Interpreter {
	return &Interpreter{
		root: &commandNode{
			"",
			make(map[string]*commandNode),
			"",
			"",
		},
	}
}

func (it *Interpreter) Run(cmd string) (err CommandError) {

	splitCmd := strings.Split(cmd, " ")
	curr := it.root
	var (
		next    *commandNode
		exisits bool
		depth   int
	)
	for depth, key := range splitCmd {

		if curr.children == nil {
			err = errors.New(fmt.Sprintf("Command node has no children: %nCommand: %s%nDepth: %d", cmd, depth))
		} else {
			next, exisits = curr.children[key]
		}

		if !exisits {
			err = InvalidCommandError{cmd, depth}
			return
		}

		curr = next
	}

	err = curr.verifyArgs(splitCmd[depth:])
	if err != nil {
		switch err.(type) {
		case InvalidArgsError:
			err = InvalidArgsError(err.Depth()
		}
	}

	return
}

func (it *Interpreter) AddCommand(path, actExpr, argsExpr string) (err error) {
	splitCmd := strings.Split(path, " ")
	curr := it.root
	var (
		next    *commandNode
		exisits bool
	)
	for _, key := range splitCmd {
		if curr.actExpr != "" {
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
				"",
				"",
			}
			curr.children[key] = next
		}
		curr = next
	}

	//if we got this far, the command is valid and the last commandNode is stored in curr
	curr.actExpr = actExpr
	curr.argExpr = argExpr
	return
}

//-----Command Tree structure------

type commandNode struct {
	key      string
	children map[string]*commandNode
	actExpr  string
	argExpr  string
}
