package commands

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type RoleCommand struct{}

func (rc *RoleCommand) Execute(pack *CommPackage) {
	server, err := db.ServerQueryOrInsert(pack.guild.ID)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an error loading server information!")
		return
	}
	var vetRole *discordgo.Role
	if server.VeteranRole.Valid {
		vetRole = util.FindRoleById(pack.guild.Roles, server.VeteranRole.String)
	}
	if len(pack.params) == 0 {
		// go find all the triggers for this server
		triggers, err := db.CustomRoleQueryServer(pack.guild.ID)
		if err != nil {
			pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue fetching the server. This is an issue with moebot!")
			return
		}
		if vetRole != nil {
			// triggers = append(triggers, vetRole.Name)
			// this should be the name of the role, but role is restricted to one word right now...
			triggers = append(triggers, "veteran")
		}
		// little hackey, but include each command in quotes
		pack.session.ChannelMessageSend(pack.channel.ID, "Possible role commands for this server: `"+strings.Join(triggers, "`, `")+"`")
	} else {
		var role *discordgo.Role
		if strings.EqualFold(pack.params[0], "veteran") {
			// before anything, if the server doesn't have a rank or role bail out
			if !server.VeteranRank.Valid || !server.VeteranRole.Valid {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, this server isn't setup to handle veteran role yet! Contact the server admins.")
				return
			}
			usr, err := db.UserServerRankQuery(pack.message.Author.ID, pack.guild.ID)
			var pointCountMessage string
			if usr != nil {
				pointCountMessage = fmt.Sprintf("%.2f%% of the way to veteran", float64(usr.Rank)/float64(server.VeteranRank.Int64)*100)
			} else {
				pointCountMessage = "Unranked"
			}
			if err != nil {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, you don't have enough veteran points yet! You're currently: "+pointCountMessage)
				return
			}
			if int64(usr.Rank) >= server.VeteranRank.Int64 {
				role = vetRole
			} else {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, you don't have enough veteran points yet! You're currently: "+pointCountMessage)
				return
			}
		} else {
			// load up the trigger to see if it exists
			roleId, err := db.CustomRoleQuery(pack.params[0], pack.guild.ID)
			if err != nil {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue fetching the role. Please provide a valid role command. Did you perhaps mean `team`, `rank`, or `NSFW`?")
				return
			}
			role = util.FindRoleById(pack.guild.Roles, roleId)
			if role == nil {
				log.Println("Nil role when searching for role id:" + roleId)
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue finding that role in this server.")
				return
			}
		}
		if util.StrContains(pack.member.Roles, role.ID, util.CaseSensitive) {
			pack.session.GuildMemberRoleRemove(pack.guild.ID, pack.message.Author.ID, role.ID)
			pack.session.ChannelMessageSend(pack.channel.ID, "Removed role "+role.Name+" for "+pack.message.Author.Mention())
		} else {
			pack.session.GuildMemberRoleAdd(pack.guild.ID, pack.message.Author.ID, role.ID)
			pack.session.ChannelMessageSend(pack.channel.ID, "Added role "+role.Name+" for "+pack.message.Author.Mention())
		}
	}
}

func (rc *RoleCommand) GetPermLevel() db.Permission {
	return db.PermAll
}

func (rc *RoleCommand) GetCommandKeys() []string {
	return []string{"ROLE"}
}

