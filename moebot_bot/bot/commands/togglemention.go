package commands

import (
	"strings"
	"time"

	"github.com/camd67/moebot/moebot_bot/util/db"

	"github.com/bwmarrin/discordgo"
)

type MentionCommand struct {
}

func (mc *MentionCommand) Execute(pack *CommPackage) {
	roleName := strings.Join(pack.params, " ")
	for _, role := range pack.guild.Roles {
		if role.Name == roleName {
			editedRole, err := pack.session.GuildRoleEdit(pack.guild.ID, role.ID, role.Name, role.Color, role.Hoist, role.Permissions, !role.Mentionable)
			if err != nil {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was a problem editing the role, try again later")
				return
			}
			go restoreMention(pack, editedRole)
			message := "Successfully changed " + editedRole.Name + " to "
			if editedRole.Mentionable {
				message += "mentionable"
			} else {
				message += "not mentionable"
			}
			pack.session.ChannelMessageSend(pack.channel.ID, message)
			return
		}
	}
	pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, could not find role "+roleName+". Please check the role name and try again.")
}

func restoreMention(pack *CommPackage, role *discordgo.Role) {
	<-time.After(5 * time.Minute)
	editedRole, err := pack.session.GuildRoleEdit(pack.guild.ID, role.ID, role.Name, role.Color, role.Hoist, role.Permissions, !role.Mentionable)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was a problem editing the role, try again later")
		return
	}
	message := "Restored role " + editedRole.Name + " to "
	if editedRole.Mentionable {
		message += "mentionable"
	} else {
		message += "not mentionable"
	}
	pack.session.ChannelMessageSend(pack.channel.ID, message)
}

func (mc *MentionCommand) GetPermLevel() db.Permission {
	return db.PermMod
}

func (mc *MentionCommand) GetCommandKeys() []string {
	return []string{"TOGGLEMENTION"}
}
