package commands

type CommandContext struct {
	
}

type Command struct {
	tokens []token
	ctx 
}

type token string 

func (t *token) 