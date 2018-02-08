package commands

import (
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

type MentionCommand struct {
	Checker *PermissionChecker
}

func (mc *MentionCommand) Execute(pack *CommPackage) {
	if !mc.Checker.HasModPerm(pack.message.Author.ID, pack.member.Roles) {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, this command has a minimum permission of mod")
		return
	}
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

func (mc *MentionCommand) Setup(session *discordgo.Session) {}

func (mc *MentionCommand) EventHandlers() []interface{} { return []interface{}{} }

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
