package discord

import (
	"errors"
	"fmt"
	"schoolbot/internal/model"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func makeChan(cls *model.Class, gc *GuildConnection) error {

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

