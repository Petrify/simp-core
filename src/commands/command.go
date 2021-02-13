package commands

type CommandContext interface {
	
}

type Command struct {
	tokens []token
	ctx 
}

type token string 

func (t *token) 