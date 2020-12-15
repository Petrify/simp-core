package discord

import (
	"errors"
	"fmt"
	"strings"

	"database/sql"

	"github.com/bwmarrin/discordgo"
	"github.com/sahilm/fuzzy"
)

type module struct {
	Id    int16
	Title string
	Abbr  string
}

type catalog []module

func (c catalog) String(i int) string {
	return fmt.Sprintf("%s %s", c[i].Abbr, c[i].Title)
}

func (c catalog) Len() int {
	return len((c))
}

func getCatalog(db *sql.DB) (c catalog) {
	rows, _ := db.Query("SELECT id, title, abbreviation FROM module")
	c = make([]module, 0, len(rows))
	for rows.next() {
		m = module{}
		rows.Scan(&m.ID, &m.Title, &m.Abbr)
		append(c, m)
	}
	return c
}

func searchCatalog(cat []catalog, key string, max int) (matches []*module) {
	results := fuzzy.FindFrom(key, cat)
	if results == nil {
		return matches
	}

	matches = make([]*model.Class, 0, max)
	for i, r := range results {
		if i >= max {
			break
		}
		matches = append(matches, clsCatalog[r.Index])
	}
	return
}

func makeChan(cls *module, gc *GuildConnection) error {

	roles, _ := gc.DS.GuildRoles(gc.guildID)
	basePerm := roles[0].Permissions //roles[0] is the @everyone role

	// check to see if the class already has a channel/role
	if cls.ChannelID != "" {
		return errors.New("Channels already exist")
	}

	chanName := fmt.Sprintf("%s [%d]", strings.TrimSpace(cls.Name), cls.Id)
	channel, err := gc.DS.GuildChannelCreate(gc.guildID, chanName, discordgo.ChannelTypeGuildText)
	if err != nil {
		return errors.New(fmt.Sprintln("Failed to create channel ", chanName, ": ", err))
	}

	role, err := gc.DS.GuildRoleCreate(gc.guildID)
	if err != nil {
		gc.DS.ChannelDelete(channel.ID)
		return errors.New(fmt.Sprint("Failed to create role for channel ", chanName, ": ", err))
	}

	gc.DS.GuildRoleEdit(gc.guildID, role.ID, chanName, 0, false, basePerm, true)

	g, err := gc.DS.Guild(gc.guildID)

	channelProp := &discordgo.ChannelEdit{
		Position: len(g.Channels),
		PermissionOverwrites: []*discordgo.PermissionOverwrite{
			newPermViewChan("", gc.guildID, false),
			newPermViewChan(role.ID, gc.guildID, true),
		},
	}

	gc.DS.ChannelEditComplex(channel.ID, channelProp)

	cls.ChannelID = channel.ID
	cls.RoleID = role.ID
	gc.box.Update(cls)
	return nil
}

func getModule(id int) (m module) {
	DB.QueryRow("SELECT id, title, abbreviation FROM module WHERE id")
	m = module{}
	return m
}
