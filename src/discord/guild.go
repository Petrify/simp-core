package discord

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type GuildManager struct {
	DS     *discordgo.Session
	active map[string]*GuildConnection
}

func NewGuildManager(token string) (gm *GuildManager, err error) {
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
		c, ok := gm.active[m.GuildID]
		if !ok {
			panic("Fuck it happened!")
		}
		c.handle(m)
	})

	// Register func as a callback for GuildCreate events.
	dg.AddHandler(func(s *discordgo.Session, g *discordgo.GuildCreate) {
		gm.newConnection(g.Guild)
	})

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages | 1)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	gm = &GuildManager{
		DS:     dg,
		active: make(map[string]*GuildConnection),
	}

	return
}

func (gm *GuildManager) newConnection(guild *discordgo.Guild) error {

	cList, err := gm.DS.GuildChannels(guild.ID)
	if err != nil {
		return err
	}

	var cmdChan *discordgo.Channel
	for _, c := range cList {
		if c.Name == "bot-commands" {
			cmdChan = c
			break
		}
	}

	if cmdChan == nil {
		cmdChan, err = gm.DS.GuildChannelCreate(guild.ID, "bot-commands", discordgo.ChannelTypeGuildText)
	}

	gc := &GuildConnection{
		DS:        gm.DS,
		cmdPrefix: "!",
		cmdChan:   cmdChan,
		guildID:   guild.ID,
		guild:     guild,
		commands:  nil,
	}

	gc.commands = gCommands()

	gm.active[guild.ID] = gc

	return nil
}

func (gm *GuildManager) Close() {
	gm.DS.Close()
}

type GuildConnection struct {
	DS        *discordgo.Session
	cmdPrefix string
	cmdChan   *discordgo.Channel
	guildID   string
	guild     *discordgo.Guild
	commands  *GuildInterpreter
}

func (gc *GuildConnection) handle(m *discordgo.MessageCreate) {
	if m.ChannelID != gc.cmdChan.ID {
		return
	}
	if strings.HasPrefix(m.Message.Content, gc.cmdPrefix) {
		err := gc.commands.Run(strings.TrimPrefix(strings.ToLower(m.Message.Content), gc.cmdPrefix), gc, m)
		if err != nil {
			gc.handleCmdErr(err)
		}
	}
}

func (gc *GuildConnection) handleCmdErr(e CommandError) {
	switch e.(type) {
	case *InvalidCommandError:
		gc.WriteCC("Unknown command")
	case *InvalidArgsError:
		gc.WriteCC("Command does not support those args") //TODO: NYI
	case *ExecutionError:
		gc.WriteCC(e.Error())
	}
}

func (gc *GuildConnection) WriteCC(text string) {
	gc.DS.ChannelMessageSend(gc.cmdChan.ID, text)
}

//Command interpreter for guilds
type GuildInterpreter struct {
	root *gNode
}

func NewGuildInterpreter() *GuildInterpreter {
	return &GuildInterpreter{
		root: &gNode{
			"",
			nil,
			nil,
		},
	}
}

func (it *GuildInterpreter) Run(cmd string, gc *GuildConnection, raw *discordgo.MessageCreate) (err CommandError) {
	splitCmd := strings.Split(cmd, " ")
	depth, err := it.root.run(splitCmd, gc, raw)
	if err != nil {
		err.setPath(splitCmd[:depth])
	}
	return
}

func (it *GuildInterpreter) AddCommand(path string, f func([]string, *GuildConnection, *discordgo.MessageCreate) error) (err error) {
	splitCmd := strings.Split(path, " ")
	curr := it.root
	var (
		next    *gNode
		exisits bool
	)
	for _, key := range splitCmd {
		if curr.function != nil {
			return errors.New("Can not create command as subcommand of an existing command")
		}

		if curr.children == nil {
			exisits = false
			curr.children = make(map[string]*gNode)
		} else {
			next, exisits = curr.children[key]
		}

		//if child does not yet exist, create one
		if !exisits {
			next = &gNode{
				key,
				nil,
				nil,
			}
			curr.children[key] = next
		}
		curr = next
	}
	//if we got this far, the command is valid and the last gNode is stored in curr
	curr.function = f
	return
}

//-----Command Tree structure------

type gNode struct {
	key      string
	children map[string]*gNode
	function func([]string, *GuildConnection, *discordgo.MessageCreate) error
}

//runs the command at ct OR recursively runs the next child command
func (ct *gNode) run(cmd []string, gc *GuildConnection, raw *discordgo.MessageCreate) (depth int, err CommandError) {
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

		depth, err = child.run(cmd[1:], gc, raw)
		depth++
		return

	}

	cmdErr := ct.function(cmd, gc, raw)
	if cmdErr != nil {
		err = &ExecutionError{
			CmdErr: cmdErr,
		}
	}

	return
}

func gCommands() (I *GuildInterpreter) {
	I = NewGuildInterpreter()
	I.AddCommand("terminal admin", adminTerminal)
	I.AddCommand("edit", classTerminal)
	return
}

func adminCommands() (I *TerminalInterpreter) {
	I = NewTerminalInterpreter()
	I.AddCommand("db add", dbAdd)
	I.AddCommand("db del", dbDel)
	I.AddCommand("db rename", dbRename)
	I.AddCommand("server clear", serverClear)
	I.AddCommand("del", chanDel)
	I.AddCommand("search", search)
	return
}

func classEditCommands() (I *TerminalInterpreter) {
	I = NewTerminalInterpreter()
	I.AddCommand("search", search)
	I.AddCommand("join", join)
	I.AddCommand("leave", clsLeave)
	return
}

func adminTerminal(args []string, gc *GuildConnection, raw *discordgo.MessageCreate) error {
	//TODO: My ID hardcoded as Amdin (bad)
	if raw.Author.ID == "84787975480700928" {
		return TManager.NewTerminal(raw.Author.ID, adminCommands(), gc, "600s", "Started an Admin Terminal")
	}
	gc.WriteCC("Access denied")
	return nil
}

func classTerminal(args []string, gc *GuildConnection, raw *discordgo.MessageCreate) error {
	return TManager.NewTerminal(raw.Author.ID, classEditCommands(), gc, "600s",
		"Hallo! Ich kann dir helfen deine Vorlesungen zu konfigurieren! Ganz einfach diese Commands (ohne `!`) eingeben.\n`search <begriff>` um nach vorlesungen zu suchen\n`join <ID>` um der Volesung beizutreten\n`leave <ID>` um eine Vorlesung zu verlassen")
}

// func dbUpdate(args []string, t *TerminalSession, raw *discordgo.MessageCreate) error {
// 	n, err := campusboard.UpdateDB(t.Origin.box)
// 	if err != nil {
// 		t.Write("Database update failed")
// 	}
// 	t.Origin.readCatalog()
// 	t.Write(fmt.Sprintf("Added %d new values to Database and Reloaded", n))
// 	return nil
// }

// func dbClear(args []string, t *TerminalSession, raw *discordgo.MessageCreate) error {
// 	t.Origin.box.RemoveAll()
// 	t.Origin.readCatalog()
// 	t.Write("Database emptied")
// 	return nil
// }

func dbDel(args []string, t *TerminalSession, raw *discordgo.MessageCreate) error {
	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return err
	}

	var module cls{}
	

	if cls == nil {
		return t.Write("No Class with this ID.")
	}

	if cls.ChannelID != "" {
		t.Origin.DS.ChannelDelete(cls.ChannelID)
		t.Origin.DS.GuildRoleDelete(t.Origin.guildID, cls.RoleID)
	}

	t.Origin.box.Remove(cls)

	return nil
}

func dbRename(args []string, t *TerminalSession, raw *discordgo.MessageCreate) error {
	if len(args) < 2 {
		return t.Write("Invalid args <int> <string>.")
	}
	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return t.Write(args[0] + " Ist keine Nummer")
	}
	cls, _ := t.Origin.box.Get(id)
	if cls == nil {
		return t.Write("Es gibt keine Volesung mit dieser ID")
	}

	name := strings.Join(args[1:], " ")
	cls.Name = name
	t.Origin.box.Update(cls)
	return nil
}

func dbAdd(args []string, t *TerminalSession, raw *discordgo.MessageCreate) error {
	name := strings.Join(args, " ")
	query := t.Origin.box.Query(model.Class_.Name.Equals(name, false))
	query.Limit(1)
	results, err := query.Find()
	if err != nil {
		return err
	} else if len(results) == 0 {
		newCls := &model.Class{
			Name:   name,
			Abbr:   name,
			Majors: make([]string, 0),
		}
		t.Origin.box.Put(newCls)
		t.Origin.readCatalog()
		return nil
	}
	return errors.New("Class already Exists")
}

func chanDel(args []string, t *TerminalSession, raw *discordgo.MessageCreate) error {

	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return err
	}
	cls, _ := t.Origin.box.Get(id)
	if cls == nil {
		return t.Write("No Class with this ID.")
	}

	if !cls.HasChan() {
		return t.Write("This class has no Channel.")
	}

	t.Origin.DS.ChannelDelete(cls.ChannelID)
	t.Origin.DS.GuildRoleDelete(t.Origin.guildID, cls.RoleID)

	cls.ChannelID = ""
	cls.RoleID = ""

	return t.Origin.box.Update(cls)
}

func newPermViewChan(roleID string, guildID string, canView bool) (perm *discordgo.PermissionOverwrite) {
	perm = new(discordgo.PermissionOverwrite)
	perm.ID = guildID //@everyone role ID is the guild's ID (this is used as the default)
	perm.Type = "role"

	if roleID != "" {
		perm.ID = roleID
	}

	if canView {
		perm.Allow = 0x00000400 // permission for view channel -- source: https://discord.com/developers/docs/topics/permissions
	} else {
		perm.Deny = 0x00000400
	}
	return
}

func serverClear(args []string, t *TerminalSession, raw *discordgo.MessageCreate) error {
	for _, cls := range t.Origin.clsCatalog {
		// check to see if the class channel/role
		if !cls.HasChan() {
			continue
		}

		t.Origin.DS.ChannelDelete(cls.ChannelID)
		t.Origin.DS.GuildRoleDelete(t.Origin.guildID, cls.RoleID)

		cls.ChannelID = ""
		cls.RoleID = ""

		t.Origin.box.Update(cls)
	}
	return nil
}

func search(args []string, t *TerminalSession, raw *discordgo.MessageCreate) error {

	var key string
	var max int
	if len(args) == 0 {
		return t.Write("No search parameters given.")
	}
	max, err := strconv.Atoi(args[0])
	if err != nil {
		max = 5
		key = strings.Join(args, " ")
	} else {
		key = strings.Join(args[1:], " ")
	}

	matches := searchCatalog(key, t.Origin.clsCatalog, max)
	if len(matches) == 0 {
		return t.Write(fmt.Sprintf("No search results found for **%s**", key))
	}
	resp := strings.Builder{}
	resp.WriteString(fmt.Sprintf("Search results for **%s**\n", key))
	resp.WriteString("```[-ID-] | Class\n")
	for _, m := range matches {
		resp.WriteString(fmt.Sprintf("[%4d] | %s\n", m.Id, m.Name))
	}
	resp.WriteString("```")
	t.Write(resp.String())
	return nil
}

func join(args []string, t *TerminalSession, raw *discordgo.MessageCreate) error {
	if len(args) == 0 {
		return t.Write("Es wurde keine ID gegeben.")
	}
	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return t.Write(args[0] + " Ist keine Nummer")
	}
	cls, _ := t.Origin.box.Get(id)
	if cls == nil {
		return t.Write("Es gibt keine Volesung mit dieser ID")
	}
	if !cls.HasChan() {
		makeChan(cls, t.Origin)
	}
	t.Origin.DS.GuildMemberRoleAdd(t.Origin.guildID, t.UserID, cls.RoleID)
	t.Write(fmt.Sprintf("%s wurde zu Deinen Vorlesungen hinzugefügt.", cls.Name))
	return nil
}

func clsLeave(args []string, t *TerminalSession, raw *discordgo.MessageCreate) error {

	if len(args) == 0 {
		return t.Write("Es wurde keine ID gegeben.")
	}
	id, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return t.Write(args[0] + " Ist keine Nummer")
	}
	cls, _ := t.Origin.box.Get(id)
	if cls == nil {
		return t.Write("Es gibt keine Volesung mit dieser ID")
	}

	t.Origin.DS.GuildMemberRoleRemove(t.Origin.guildID, t.UserID, cls.RoleID)
	return t.Write(fmt.Sprintf("Du hast die Vorlesung %s verlassen.", cls.Name))
}

func test(args []string, t *TerminalSession, raw *discordgo.MessageCreate) error {
	t.Write(t.Origin.guild.Members[0].Nick + ":")
	t.Write(t.Origin.guild.Members[0].Roles[0])
	return nil
}
