package commands

import (
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type RoleCommand struct{}

func (rc RoleCommand) Execute(pack *CommPackage) {
	server, err := db.ServerQueryOrInsert(pack.Guild.ID)
	if err != nil {
		pack.Session.ChannelMessageSend(pack.Channel.ID, "Sorry, there was an error loading server information!")
		return
	}
	var vetRole *discordgo.Role
	if server.VeteranRole.Valid {
		vetRole = util.FindRoleById(pack.Guild.Roles, server.VeteranRole.String)
	}
	if len(pack.Params) == 0 {
		// go find all the triggers for this server
		triggers, err := db.CustomRoleQueryServer(pack.Guild.ID)
		if err != nil {
			pack.Session.ChannelMessageSend(pack.Channel.ID, "Sorry, there was an issue fetching the server. This is an issue with moebot!")
			return
		}
		if vetRole != nil {
			// triggers = append(triggers, vetRole.Name)
			// this should be the name of the role, but role is restricted to one word right now...
			triggers = append(triggers, "veteran")
		}
		// little hackey, but include each command in quotes
		pack.Session.ChannelMessageSend(pack.Channel.ID, "Possible role commands for this server: `"+strings.Join(triggers, "`, `")+"`")
	} else {
		var role *discordgo.Role
		if strings.EqualFold(pack.Params[0], "veteran") {
			// before anything, if the server doesn't have a rank or role bail out
			if !server.VeteranRank.Valid || !server.VeteranRole.Valid {
				pack.Session.ChannelMessageSend(pack.Channel.ID, "Sorry, this server isn't setup to handle veteran role yet! Contact the server admins.")
				return
			}
			usr, err := db.UserServerRankQuery(pack.Message.Author.ID, pack.Guild.ID)
			var pointCountMessage string
			if usr != nil {
				pointCountMessage = strconv.Itoa(usr.Rank)
			} else {
				pointCountMessage = "Unranked"
			}
			if err != nil {
				pack.Session.ChannelMessageSend(pack.Channel.ID, "Sorry, you don't have enough veteran points yet! Your current points: "+pointCountMessage+
					" Points required for veteran: "+strconv.Itoa(int(server.VeteranRank.Int64)))
				return
			}
			if int64(usr.Rank) >= server.VeteranRank.Int64 {
				role = vetRole
			} else {
				pack.Session.ChannelMessageSend(pack.Channel.ID, "Sorry, you don't have enough veteran points yet! Your current points: "+pointCountMessage+
					" Points required for veteran: "+strconv.Itoa(int(server.VeteranRank.Int64)))
				return
			}
		} else {
			// load up the trigger to see if it exists
			roleId, err := db.CustomRoleQuery(pack.Params[0], pack.Guild.ID)
			if err != nil {
				pack.Session.ChannelMessageSend(pack.Channel.ID, "Sorry, there was an issue fetching the role. Please provide a valid role command. Did you perhaps mean `team`, `rank`, or `NSFW`?")
				return
			}
			role = util.FindRoleById(pack.Guild.Roles, roleId)
			if role == nil {
				log.Println("Nil role when searching for role id:" + roleId)
				pack.Session.ChannelMessageSend(pack.Channel.ID, "Sorry, there was an issue finding that role in this server.")
				return
			}
		}
		if util.StrContains(pack.Member.Roles, role.ID, util.CaseSensitive) {
			pack.Session.GuildMemberRoleRemove(pack.Guild.ID, pack.Message.Author.ID, role.ID)
			pack.Session.ChannelMessageSend(pack.Channel.ID, "Removed role "+role.Name+" for "+pack.Message.Author.Mention())
		} else {
			pack.Session.GuildMemberRoleAdd(pack.Guild.ID, pack.Message.Author.ID, role.ID)
			pack.Session.ChannelMessageSend(pack.Channel.ID, "Added role "+role.Name+" for "+pack.Message.Author.Mention())
		}
	}
}
