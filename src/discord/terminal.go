package discord

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

const closeCommand string = "!close"
const killCommand string = "!kill"

type TerminalManager struct {
	DiscordSession *discordgo.Session
	activeTerms    map[string]*TerminalSession //Mapped by channelID
	toClose        chan string                 //recieves channelIDs that are to be closed
}

func NewTerminalManager(token string) (tm *TerminalManager, err error) {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Register func as a callback for MessageCreate events.
	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {

		//first, make sure the bot can't reply to itself
		if m.Author.Bot {
			return
		}
		term, ok := tm.activeTerms[m.ChannelID]
		if !ok {
			s.ChannelMessageSend(m.ChannelID, "There is currently no active terminal on this channel. Please go to your SimpBot administered server to start a new terminal")
			return
		}
		switch m.Message.Content {
		case closeCommand:
			term.Close("Close command issued")
		case killCommand:
			tm.removeTerm(m.ChannelID)
			s.ChannelMessageSend(m.ChannelID, "Terminal killed.")
		default:
			term.in <- m
		}

	})

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsDirectMessages)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	tm = &TerminalManager{
		DiscordSession: dg,
		activeTerms:    make(map[string]*TerminalSession),
	}

	return
}

func (tm *TerminalManager) NewTerminal(userID string, I *TerminalInterpreter, origin *GuildConnection, timeout string, welcome string) (err error) {

	channel, err := tm.DiscordSession.UserChannelCreate(userID)
	_, ok := tm.activeTerms[channel.ID]
	if ok {
		tm.DiscordSession.ChannelMessageSend(channel.ID, fmt.Sprintf("There is already an active terminal on this channel. Please use %s to close this terminal before opening a new one. If the terminal is stuck, use %s (not recommended)", closeCommand, killCommand))
		return
	}

	t, err := time.ParseDuration(timeout)
	if err != nil {
		return err
	}

	term := &TerminalSession{
		manager: tm,
		DS:      tm.DiscordSession,
		UserID:  userID,
		ChanID:  channel.ID,
		Origin:  origin,

		in:   make(chan *discordgo.MessageCreate),
		stop: make(chan string, 1),

		commands: I,

		tMax:  t,
		timer: time.NewTimer(t),
	}
	term.timer.Stop() //so that the timer only truly starts when the terminal's loop begins
	tm.activeTerms[channel.ID] = term
	term.Write(welcome)
	go term.loop()
	return
}

func (tm *TerminalManager) Close() {
	for _, term := range tm.activeTerms {
		term.Close("Terminal manager is shutting down.")
	}
	tm.DiscordSession.Close()
}

//Removes a terminal from the tm
func (tm *TerminalManager) removeTerm(ID string) {
	_, ok := tm.activeTerms[ID]
	if ok {
		delete(tm.activeTerms, ID)
	}

}

// Terminal Session
type TerminalSession struct {
	manager *TerminalManager

	DS     *discordgo.Session
	UserID string
	ChanID string
	Origin *GuildConnection

	in   chan *discordgo.MessageCreate
	stop chan string

	commands *TerminalInterpreter

	tMax  time.Duration
	timer *time.Timer
}

func (t *TerminalSession) Write(text string) (err error) {
	_, err = t.DS.ChannelMessageSend(t.ChanID, text)
	return
}

//Read:
//Reads the next user input to the Terminal. Blocks until there is an input available or the terminal is closed
func (t *TerminalSession) Read() (text string, ok bool) {
	msg, ok := <-t.in
	if !ok {
		return "", ok
	}
	text = msg.Message.Content
	return
}

func (t *TerminalSession) loop() {
	var inp *discordgo.MessageCreate
	for {

		//force checking timer & stop signal before allowing checking input
		select {
		case <-t.timer.C:
			t.cleanup("The session has expired")
			return

		case reason := <-t.stop:
			t.cleanup(reason)
			return

		default:
			select {
			case reason := <-t.stop:
				t.cleanup(reason)
				return

			case <-t.timer.C:
				t.cleanup("The session has expired")
				return

			case inp = <-t.in:
				t.timer.Stop() //so that the session does not expire during command execution
				err := t.commands.Run(strings.ToLower(inp.Message.Content), t, inp)
				if err != nil {
					t.handleCmdErr(err)
				}
			}
		}
		t.timer.Reset(t.tMax)
	}
}

func (t *TerminalSession) handleCmdErr(e CommandError) {
	switch e.(type) {
	case *InvalidCommandError:
		t.Write("Unknown command")
	case *InvalidArgsError:
		t.Write("Command does not support those args") //TODO: NYI
	case *ExecutionError:
		t.Write(e.Error())
	}
}

func (t *TerminalSession) Close(reason string) {
	t.stop <- reason
}

func (t *TerminalSession) cleanup(reason string) {
	t.timer.Stop()
	t.manager.removeTerm(t.ChanID)
	close(t.in)
	t.Write(fmt.Sprintf("Terminal is now closed.\nReason: %s\n", reason))
}

//TerminalInterpreter:
//similar to a normal command interpreter except that it operates on a terminal instead
type TerminalInterpreter struct {
	root *tNode
}

func NewTerminalInterpreter() *TerminalInterpreter {
	return &TerminalInterpreter{
		root: &tNode{
			"",
			nil,
			nil,
		},
	}
}

func (it *TerminalInterpreter) Run(cmd string, t *TerminalSession, raw *discordgo.MessageCreate) (err CommandError) {
	splitCmd := strings.Split(cmd, " ")
	depth, err := it.root.run(splitCmd, t, raw)
	if err != nil {
		err.setPath(splitCmd[:depth])
	}
	return
}

func (it *TerminalInterpreter) AddCommand(path string, f func(args []string, t *TerminalSession, raw *discordgo.MessageCreate) error) (err error) {
	splitCmd := strings.Split(path, " ")
	curr := it.root
	var (
		next    *tNode
		exisits bool
	)
	for _, key := range splitCmd {
		if curr.function != nil {
			return errors.New("Can not create command as subcommand of an existing command")
		}

		if curr.children == nil {
			exisits = false
			curr.children = make(map[string]*tNode)
		} else {
			next, exisits = curr.children[key]
		}

		//if child does not yet exist, create one
		if !exisits {
			next = &tNode{
				key,
				nil,
				nil,
			}
			curr.children[key] = next
		}
		curr = next
	}
	//if we got this far, the command is valid and the last tNode is stored in curr
	curr.function = f
	return
}

//-----Command Tree structure------

type tNode struct {
	key      string
	children map[string]*tNode
	function func(args []string, t *TerminalSession, raw *discordgo.MessageCreate) error
}

//runs the command at ct OR recursively runs the next child command
func (ct *tNode) run(cmd []string, t *TerminalSession, raw *discordgo.MessageCreate) (depth int, err CommandError) {
	if ct.function == nil {

		if len(cmd) == 0 {
			err = &InvalidCommandError{}
			return
		}

		nextWord := cmd[0]
		child, ok := ct.children[nextWord]
		if !ok {
			err = &InvalidCommandError{}
			return
		}

		depth, err = child.run(cmd[1:], t, raw)
		depth++
		return

	}

	cmdErr := ct.function(cmd, t, raw)
	if cmdErr != nil {
		err = &ExecutionError{
			CmdErr: cmdErr,
		}
	}

	return
}
