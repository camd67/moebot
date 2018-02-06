package bot

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/camd67/moebot/moebot_bot/bot/commands"
	"github.com/camd67/moebot/moebot_bot/util/db"

	"github.com/camd67/moebot/moebot_bot/util"

	"github.com/bwmarrin/discordgo"
)

//Moved to bot.go
//var commandsMap = map[string]func(pack *commPackage){
// "TEAM":          commTeam,
// "ROLE":          commRole,
// "RANK":          commRank,
// "NSFW":          commNsfw,
// "HELP":          commHelp,
// "CHANGELOG":     commChange,
// "RAFFLE":        commRaffle,
// "SUBMIT":        commSubmit,
// "ECHO":          commEcho,
// "PERMIT":        commPermit,
// "CUSTOM":        commCustom,
// "PING":          commPing,
// "SPOILER":       commSpoiler,
// "POLL": 			commPoll,
// "TOGGLEMENTION": commToggleMention,
// "SERVER":        commServer,
// "PROFILE":       commProfile,
// "PINMOVE":       commPinMove,
//}

func RunCommand(session *discordgo.Session, message *discordgo.Message, guild *discordgo.Guild, channel *discordgo.Channel, member *discordgo.Member) {
	messageParts := strings.Split(message.Content, " ")
	if len(messageParts) <= 1 {
		// bad command, missing command after prefix
		return
	}
	command := strings.ToUpper(messageParts[1])

	if commFunc, commPresent := commandsMap[command]; commPresent {
		params := messageParts[2:]
		log.Println("Processing command: " + command + " from user: {" + fmt.Sprintf("%+v", message.Author) + "}| With Params:{" + strings.Join(params, ",") + "}")
		session.ChannelTyping(message.ChannelID)
		commFunc.Execute(&commands.CommPackage{session, message, guild, member, channel, params})
	}
}

func commProfile(pack *commands.CommPackage) {
	// technically we'll already have a user + server at this point, but may not have a usr. Still create if necessary
	_, err := db.ServerQueryOrInsert(pack.Guild.ID)
	_, err = db.UserQueryOrInsert(pack.Message.Author.ID)
	usr, err := db.UserServerRankQuery(pack.Message.Author.ID, pack.Guild.ID)
	if err != nil {
		if err != sql.ErrNoRows {
			pack.Session.ChannelMessageSend(pack.Message.ChannelID, "Sorry, there was an issue getting your information!")
			return
		}
		if err == sql.ErrNoRows || usr == nil {
			pack.Session.ChannelMessageSend(pack.Message.ChannelID, "Sorry, you don't have a rank yet!")
			return
		}
	}
	pack.Session.ChannelMessageSend(pack.Message.ChannelID, util.UserIdToMention(pack.Message.Author.ID)+"'s profile:\nRank: "+strconv.Itoa(usr.Rank))
}

func commServer(pack *commands.CommPackage) {
	// if m := HasModPerm(pack.Message.Author.ID, pack.Member.Roles); !m {
	// 	pack.Session.ChannelMessageSend(pack.Channel.ID, "Sorry, this command has a minimum permission of mod")
	// 	return
	// }
	const possibleConfigMessages = "Possible configs: {VeteranRank -> number}, {VeteranRole -> full role name}, {BotChannel -> channel ID}"
	s, err := db.ServerQueryOrInsert(pack.Guild.ID)

	if len(pack.Params) <= 1 {
		rank := int(util.GetInt64OrDefault(s.VeteranRank))
		role := util.GetStringOrDefault(s.VeteranRole)
		if role != "unknown" {
			role = util.FindRoleById(pack.Guild.Roles, role).Name
		}
		rule := util.GetStringOrDefault(s.RuleAgreement)
		welcome := util.GetStringOrDefault(s.WelcomeMessage)
		botChannel := util.GetStringOrDefault(s.BotChannel)
		pack.Session.ChannelMessageSend(pack.Message.ChannelID, "This server's configs: {Rank: "+strconv.Itoa(rank)+"} {Role: "+role+"} {Welcome: "+welcome+
			"} {Rule Confirm: "+rule+"} {BotChannel ID: "+botChannel+"}")
		return
	}
	configKey := strings.ToUpper(pack.Params[0])
	configValue := strings.Join(pack.Params[1:], " ")
	if err != nil {
		log.Println("Error trying to get server", err)
		pack.Session.ChannelMessageSend(pack.Message.ChannelID, "Sorry, there was an error getting the server")
		return
	}
	if configKey == "VETERANRANK" {
		rank, err := strconv.Atoi(configValue)
		if err != nil || rank < 0 {
			// don't bother logging this one. Someone's just given a non-number
			pack.Session.ChannelMessageSend(pack.Message.ChannelID, "Please provide a positive number for the veteran rank")
			return
		}
		s.VeteranRank = sql.NullInt64{
			Int64: int64(rank),
			Valid: true,
		}
	} else if configKey == "VETERANROLE" {
		role := util.FindRoleByName(pack.Guild.Roles, configValue)
		if role == nil {
			pack.Session.ChannelMessageSend(pack.Message.ChannelID, "Please provide a valid role and make sure it's the full role name")
			return
		}
		s.VeteranRole = sql.NullString{
			String: role.ID,
			Valid:  true,
		}
	} else if configKey == "BOTCHANNEL" {
		c, err := pack.Session.Channel(configValue)
		if err != nil || c.Type != discordgo.ChannelTypeGuildText {
			pack.Session.ChannelMessageSend(pack.Message.ChannelID, "Please provide a valid text channel ID")
			return
		}
		s.BotChannel = sql.NullString{
			String: c.ID,
			Valid:  true,
		}
	} else {
		pack.Session.ChannelMessageSend(pack.Message.ChannelID, possibleConfigMessages)
		return
	}
	err = db.ServerFullUpdate(s)
	if err != nil {
		pack.Session.ChannelMessageSend(pack.Message.ChannelID, "Sorry, there was an error updating the server table. Your change was probably not applied.")
		return
	}
	pack.Session.ChannelMessageSend(pack.Message.ChannelID, "Updated this server!")
}

func commPermit(pack *commands.CommPackage) {
	// if m := hasValidMasterId(pack); !m {
	// 	return
	// }
	// should always have more than 2 params: permission level, role name ... role name
	if len(pack.Params) < 2 {
		pack.Session.ChannelMessageSend(pack.Message.ChannelID, "Please provide a permission level followed by the role name")
		return
	}
	permLevel := db.GetPermissionFromString(pack.Params[0])
	if permLevel == -1 {
		pack.Session.ChannelMessageSend(pack.Message.ChannelID, "Invalid permission level")
		return
	}
	// find the correct role
	roleName := strings.Join(pack.Params[1:], " ")
	r := util.FindRoleByName(pack.Guild.Roles, roleName)
	if r == nil {
		pack.Session.ChannelMessageSend(pack.Message.ChannelID, "Unknown role name")
	}
	// we've got the role, add it to the db, updating if necessary
	// but first grab the server (probably want to move this out to include in the commPackage
	s, err := db.ServerQueryOrInsert(pack.Guild.ID)
	if err != nil {
		pack.Session.ChannelMessageSend(pack.Message.ChannelID, "Error retrieving server information")
		return
	}
	db.RoleInsertOrUpdate(db.Role{
		ServerId:   s.Id,
		RoleUid:    r.ID,
		Permission: permLevel,
	})
	pack.Session.ChannelMessageSend(pack.Channel.ID, "Set permission ("+db.SprintPermission(permLevel)+") level for role "+roleName)
}

func commCustom(pack *commands.CommPackage) {
	// if !HasModPerm(pack.Message.Author.ID, pack.Member.Roles) {
	// 	pack.Session.ChannelMessageSend(pack.Channel.ID, "Sorry, this command has a minimum permission of mod")
	// 	return
	// }
	// should have params: command name - role name or delete - id
	if len(pack.Params) < 2 {
		pack.Session.ChannelMessageSend(pack.Channel.ID, "Please provide command name followed by the role name")
		return
	}
	if strings.ToUpper(pack.Params[0]) == "DELETE" {
		id, err := strconv.Atoi(pack.Params[1])
		if err != nil {
			pack.Session.ChannelMessageSend(pack.Channel.ID, "Please provide a valid ID to delete")
			return
		}
		count := db.CustomRoleDelete(id)
		if count == -1 {
			pack.Session.ChannelMessageSend(pack.Channel.ID, "Sorry, there was an issue deleting that custom role. Perhaps it doesn't exist?")
		} else {
			pack.Session.ChannelMessageSend(pack.Channel.ID, "Deleted "+strconv.FormatInt(count, 10)+" custom role commands")
		}
	} else {
		// get the role and server
		server, err := db.ServerQueryOrInsert(pack.Guild.ID)
		if err != nil {
			pack.Session.ChannelMessageSend(pack.Channel.ID, "Error storing server information")
			return
		}
		roleName := strings.Join(pack.Params[1:], " ")
		r := util.FindRoleByName(pack.Guild.Roles, roleName)
		if r == nil {
			pack.Session.ChannelMessageSend(pack.Channel.ID, "Sorry, it doesn't seem like that role exists on this server.")
			return
		}
		role, err := db.RoleQueryOrInsert(db.Role{
			ServerId: server.Id,
			RoleUid:  r.ID,
		})
		oldId, exists := db.CustomRoleRowExists(pack.Params[0], server.GuildUid)
		if !exists {
			err = db.CustomRoleAdd(pack.Params[0], server.Id, role.Id)
			if err != nil {
				pack.Session.ChannelMessageSend(pack.Channel.ID, "Error adding custom role")
				return
			}
			pack.Session.ChannelMessageSend(pack.Channel.ID, "Added custom command `"+pack.Params[0]+"` tied to the role: "+roleName)
		} else {
			pack.Session.ChannelMessageSend(pack.Channel.ID, "Custom command already exists. Delete with `"+ComPrefix+" custom delete "+strconv.Itoa(oldId)+"`")
		}
	}
}

func commEcho(pack *commands.CommPackage) {
	// if m := hasValidMasterId(pack); !m {
	// 	return
	// }
	_, err := strconv.Atoi(pack.Params[0])
	if err != nil {
		pack.Session.ChannelMessageSend(pack.Message.ChannelID, "Sorry, that's an invalid channel ID")
		return
	}
	pack.Session.ChannelMessageSend(pack.Params[0], strings.Join(pack.Params[1:], " "))
}

func commPing(pack *commands.CommPackage) {
	// seems this has some time drift when using docker for windows... need to verify if it's accurate on the server
	messageTime, _ := pack.Message.Timestamp.Parse()
	pingTime := time.Duration(time.Now().UnixNano() - messageTime.UnixNano())
	pack.Session.ChannelMessageSend(pack.Channel.ID, "Latency to server: "+pingTime.String())
}

func commRole(pack *commands.CommPackage) {
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

func commTeam(pack *commands.CommPackage) {

}

func commRank(pack *commands.CommPackage) {
}

func commNsfw(pack *commands.CommPackage) {
}

func commSpoiler(pack *commands.CommPackage) {
	content := pack.Message.Author.Mention() + " sent a spoiler"
	for i := 0; i < 2; i++ {
		err := pack.Session.ChannelMessageDelete(pack.Channel.ID, pack.Message.ID)
		if err == nil {
			break
		}
		log.Println("Error while deleting message", err)
	}

	spoilerTitle, spoilerText := util.GetSpoilerContents(pack.Params)
	if spoilerTitle != "" {
		content += ": **" + spoilerTitle + "**"
	}
	spoilerGif := util.MakeGif(spoilerText)
	pack.Session.ChannelMessageSendComplex(pack.Channel.ID, &discordgo.MessageSend{
		Content: content,
		File: &discordgo.File{
			Name:        "Spoiler.gif",
			ContentType: "image/gif",
			Reader:      bytes.NewReader(spoilerGif),
		},
	})
}

func commToggleMention(pack *commands.CommPackage) {
	// if !HasModPerm(pack.Message.Author.ID, pack.Member.Roles) {
	// 	pack.Session.ChannelMessageSend(pack.Channel.ID, "Sorry, this command has a minimum permission of mod")
	// 	return
	// }
	roleName := strings.Join(pack.Params, " ")
	for _, role := range pack.Guild.Roles {
		if role.Name == roleName {
			editedRole, err := pack.Session.GuildRoleEdit(pack.Guild.ID, role.ID, role.Name, role.Color, role.Hoist, role.Permissions, !role.Mentionable)
			if err != nil {
				pack.Session.ChannelMessageSend(pack.Channel.ID, "Sorry, there was a problem editing the role, try again later")
				return
			}
			go restoreMention(pack, editedRole)
			message := "Successfully changed " + editedRole.Name + " to "
			if editedRole.Mentionable {
				message += "mentionable"
			} else {
				message += "not mentionable"
			}
			pack.Session.ChannelMessageSend(pack.Channel.ID, message)
			return
		}
	}
	pack.Session.ChannelMessageSend(pack.Channel.ID, "Sorry, could not find role "+roleName+". Please check the role name and try again.")
}

func restoreMention(pack *commands.CommPackage, role *discordgo.Role) {
	<-time.After(5 * time.Minute)
	editedRole, err := pack.Session.GuildRoleEdit(pack.Guild.ID, role.ID, role.Name, role.Color, role.Hoist, role.Permissions, !role.Mentionable)
	if err != nil {
		pack.Session.ChannelMessageSend(pack.Channel.ID, "Sorry, there was a problem editing the role, try again later")
		return
	}
	message := "Restored role " + editedRole.Name + " to "
	if editedRole.Mentionable {
		message += "mentionable"
	} else {
		message += "not mentionable"
	}
	pack.Session.ChannelMessageSend(pack.Channel.ID, message)
}
