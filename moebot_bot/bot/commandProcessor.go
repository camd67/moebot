package bot

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/camd67/moebot/moebot_bot/util/db"

	"github.com/camd67/moebot/moebot_bot/util"

	"github.com/bwmarrin/discordgo"
)

var commands = map[string]func(pack *commPackage){
	"TEAM":          commTeam,
	"ROLE":          commRole,
	"RANK":          commRank,
	"NSFW":          commNsfw,
	"HELP":          commHelp,
	"CHANGELOG":     commChange,
	"RAFFLE":        commRaffle,
	"SUBMIT":        commSubmit,
	"ECHO":          commEcho,
	"PERMIT":        commPermit,
	"CUSTOM":        commCustom,
	"PING":          commPing,
	"SPOILER":       commSpoiler,
	"POLL":          commPoll,
	"TOGGLEMENTION": commToggleMention,
}

func RunCommand(session *discordgo.Session, message *discordgo.Message, guild *discordgo.Guild, channel *discordgo.Channel, member *discordgo.Member) {
	messageParts := strings.Split(message.Content, " ")
	if len(messageParts) <= 1 {
		// bad command, missing command after prefix
		return
	}
	command := strings.ToUpper(messageParts[1])

	if commFunc, commPresent := commands[command]; commPresent {
		var buf bytes.Buffer
		params := messageParts[2:]
		buf.WriteString("Processing command: ")
		buf.WriteString(command)
		buf.WriteString(" from user: {")
		buf.WriteString(fmt.Sprintf("%+v", message.Author))
		buf.WriteString("}| With Params:{")
		for _, p := range params {
			buf.WriteString(p)
			buf.WriteString(",")
		}
		buf.WriteString("}")
		log.Println(buf.String())
		session.ChannelTyping(message.ChannelID)
		commFunc(&commPackage{session, message, guild, member, channel, params})
	}
}

func commPermit(pack *commPackage) {
	if m := checkValidMasterId(pack); !m {
		return
	}
	// should always have more than 2 params: permission level, role name ... role name
	if len(pack.params) < 2 {
		pack.session.ChannelMessageSend(pack.message.ChannelID, "Please provide a permission level followed by the role name")
		return
	}
	permLevel := db.GetPermissionFromString(pack.params[0])
	if permLevel == -1 {
		pack.session.ChannelMessageSend(pack.message.ChannelID, "Invalid permission level")
		return
	}
	// find the correct role
	roleName := strings.Join(pack.params[1:], " ")
	r := util.FindRoleByName(pack.guild.Roles, roleName)
	if r == nil {
		pack.session.ChannelMessageSend(pack.message.ChannelID, "Unknown role name")
	}
	// we've got the role, add it to the db, updating if necessary
	// but first grab the server (probably want to move this out to include in the commPackage
	s, err := db.ServerQueryOrInsert(pack.guild.ID)
	if err != nil {
		pack.session.ChannelMessageSend(pack.message.ChannelID, "Error retrieving server information")
		return
	}
	db.RoleInsertOrUpdate(db.Role{
		ServerId:   s.Id,
		RoleUid:    r.ID,
		Permission: permLevel,
	})
	pack.session.ChannelMessageSend(pack.channel.ID, "Set permission ("+db.SprintPermission(permLevel)+") level for role "+roleName)
}

func commCustom(pack *commPackage) {
	if !HasModPerm(pack.message.Author.ID, pack.member.Roles) {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, this command has a minimum permission of mod")
		return
	}
	// should have params: command name - role name or delete - id
	if len(pack.params) < 2 {
		pack.session.ChannelMessageSend(pack.channel.ID, "Please provide command name followed by the role name")
		return
	}
	if strings.ToUpper(pack.params[0]) == "DELETE" {
		id, err := strconv.Atoi(pack.params[1])
		if err != nil {
			pack.session.ChannelMessageSend(pack.channel.ID, "Please provide a valid ID to delete")
			return
		}
		count := db.CustomRoleDelete(id)
		if count == -1 {
			pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue deleting that custom role. Perhaps it doesn't exist?")
		} else {
			pack.session.ChannelMessageSend(pack.channel.ID, "Deleted "+strconv.FormatInt(count, 10)+" custom role commands")
		}
	} else {
		// get the role and server
		server, err := db.ServerQueryOrInsert(pack.guild.ID)
		if err != nil {
			pack.session.ChannelMessageSend(pack.channel.ID, "Error storing server information")
			return
		}
		roleName := strings.Join(pack.params[1:], " ")
		r := util.FindRoleByName(pack.guild.Roles, roleName)
		if r == nil {
			pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, it doesn't seem like that role exists on this server.")
			return
		}
		role, err := db.RoleQueryOrInsert(db.Role{
			ServerId: server.Id,
			RoleUid:  r.ID,
		})
		oldId, exists := db.CustomRoleRowExists(pack.params[0], server.GuildUid)
		if !exists {
			err = db.CustomRoleAdd(pack.params[0], server.Id, role.Id)
			if err != nil {
				pack.session.ChannelMessageSend(pack.channel.ID, "Error adding custom role")
				return
			}
			pack.session.ChannelMessageSend(pack.channel.ID, "Added custom command `"+pack.params[0]+"` tied to the role: "+roleName)
		} else {
			pack.session.ChannelMessageSend(pack.channel.ID, "Custom command already exists. Delete with `"+ComPrefix+" custom delete "+strconv.Itoa(oldId)+"`")
		}
	}
}

func commEcho(pack *commPackage) {
	if m := checkValidMasterId(pack); !m {
		return
	}
	_, err := strconv.Atoi(pack.params[0])
	if err != nil {
		pack.session.ChannelMessageSend(pack.message.ChannelID, "Sorry, that's an invalid channel ID")
		return
	}
	pack.session.ChannelMessageSend(pack.params[0], strings.Join(pack.params[1:], " "))
}

func commPing(pack *commPackage) {
	// seems this has some time drift when using docker for windows... need to verify if it's accurate on the server
	messageTime, _ := pack.message.Timestamp.Parse()
	pingTime := time.Duration(time.Now().UnixNano() - messageTime.UnixNano())
	pack.session.ChannelMessageSend(pack.channel.ID, "Latency to server: "+pingTime.String())
}

func commRole(pack *commPackage) {
	if len(pack.params) == 0 {
		// go find all the triggers for this server
		triggers, err := db.CustomRoleQueryServer(pack.guild.ID)
		if err != nil {
			pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue fetching the server. This is an issue with moebot!")
			return
		}
		// little hackey, but include each command in quotes
		pack.session.ChannelMessageSend(pack.channel.ID, "Possible role commands for this server: `"+strings.Join(triggers, "`, `")+"`")
	} else {
		// load up the trigger to see if it exists
		roleId, err := db.CustomRoleQuery(pack.params[0], pack.guild.ID)
		if err != nil {
			pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue fetching the role. Please provide a valid role command. Did you perhaps mean `team` or `rank`?")
			return
		}
		role := util.FindRoleById(pack.guild.Roles, roleId)
		if role == nil {
			log.Println("Nil role when searching for role id:" + roleId)
			pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue finding that role in this server.")
			return
		}
		if util.StrContains(pack.member.Roles, roleId, util.CaseSensitive) {
			pack.session.GuildMemberRoleRemove(pack.guild.ID, pack.message.Author.ID, roleId)
			pack.session.ChannelMessageSend(pack.channel.ID, "Removed role "+role.Name+" for "+pack.message.Author.Mention())
		} else {
			pack.session.GuildMemberRoleAdd(pack.guild.ID, pack.message.Author.ID, roleId)
			pack.session.ChannelMessageSend(pack.channel.ID, "Added role "+role.Name+" for "+pack.message.Author.Mention())
		}
	}
}

func commTeam(pack *commPackage) {
	processGuildRole([]string{"Nanachi", "Ozen", "Bondrewd", "Reg", "Riko", "Maruruk"}, pack.session, pack.params, pack.channel, pack.guild, pack.message, false)
}

func commRank(pack *commPackage) {
	processGuildRole([]string{"Red", "Blue"}, pack.session, pack.params, pack.channel, pack.guild, pack.message, true)
}

func commNsfw(pack *commPackage) {
	// force NSFW comm param so we can reuse guild role
	processGuildRole([]string{"NSFW"}, pack.session, []string{"NSFW"}, pack.channel, pack.guild, pack.message, false)
}

func commSpoiler(pack *commPackage) {
	content := pack.message.Author.Mention() + " sent a spoiler"
	pack.session.ChannelMessageDelete(pack.channel.ID, pack.message.ID)
	spoilerTitle, spoilerText := util.GetSpoilerContents(pack.params)
	if spoilerTitle != "" {
		content += ": **" + spoilerTitle + "**"
	}
	spoilerGif := util.MakeGif(spoilerText)
	pack.session.ChannelMessageSendComplex(pack.channel.ID, &discordgo.MessageSend{
		Content: content,
		File: &discordgo.File{
			Name:        "Spoiler.gif",
			ContentType: "image/gif",
			Reader:      bytes.NewReader(spoilerGif),
		},
	})
}

func commPoll(pack *commPackage) {
	if !HasModPerm(pack.message.Author.ID, pack.member.Roles) {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, this command has a minimum permission of mod")
		return
	}
	if pack.params[0] == "-close" {
		pollsHandler.closePoll(pack)
		return
	}
	pollsHandler.openPoll(pack)
}

func commToggleMention(pack *commPackage) {
	if !HasModPerm(pack.message.Author.ID, pack.member.Roles) {
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

func restoreMention(pack *commPackage, role *discordgo.Role) {
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

/*
Processes a guild role based on a list of allowed role names, and a requested role.
If StrictRole is true then the first role in the list of allowed roles is used if all roles are removed.
*/
func processGuildRole(allowedRoles []string, session *discordgo.Session, params []string, channel *discordgo.Channel, guild *discordgo.Guild, message *discordgo.Message, strictRole bool) {
	if len(params) == 0 || !util.StrContains(allowedRoles, params[0], util.CaseInsensitive) {
		session.ChannelMessageSend(channel.ID, message.Author.Mention()+" please provide one of the approved roles: "+strings.Join(allowedRoles, ", ")+". Did you perhaps mean `team` or `rank`?")
		return
	}

	// get the list of roles and find the one that matches the text
	var roleToAdd *discordgo.Role
	var allRolesToChange []string
	requestedRole := strings.ToUpper(params[0])
	for _, role := range guild.Roles {
		for _, roleToCheck := range allowedRoles {
			if strings.HasPrefix(strings.ToUpper(role.Name), strings.ToUpper(roleToCheck)) {
				allRolesToChange = append(allRolesToChange, role.ID)
			}
		}
		if strings.HasPrefix(strings.ToUpper(role.Name), requestedRole) {
			roleToAdd = role
		}
	}
	if roleToAdd == nil {
		log.Println("Unable to find role", message)
		session.ChannelMessageSend(channel.ID, "Sorry "+message.Author.Mention()+" I don't recognize that role...")
		return
	}
	guildMember, err := session.GuildMember(guild.ID, message.Author.ID)
	if err != nil {
		log.Println("Unable to find guild member", err)
		return
	}
	memberRoles := guildMember.Roles
	if strictRole && util.StrContains(memberRoles, roleToAdd.ID, util.CaseSensitive) && strings.HasPrefix(strings.ToUpper(roleToAdd.Name), requestedRole) {
		// we're in strict mode, and got a removal for the first role. prevent that
		session.ChannelMessageSend(channel.ID, "You've already got that role! You can change roles but can't remove them with this command.")
		return
	}
	removeAllRoles(session, guildMember, allRolesToChange, guild)
	if util.StrContains(memberRoles, roleToAdd.ID, util.CaseSensitive) {
		session.GuildMemberRoleRemove(guild.ID, message.Author.ID, roleToAdd.ID)
		session.ChannelMessageSend(channel.ID, "Removed role: "+roleToAdd.Name+" for "+message.Author.Mention())
		log.Println("Removed role " + roleToAdd.Name + " to user: " + message.Author.Username)
	} else {
		session.GuildMemberRoleAdd(guild.ID, message.Author.ID, roleToAdd.ID)
		session.ChannelMessageSend(channel.ID, "Added role: "+roleToAdd.Name+" for "+message.Author.Mention())
		log.Println("Added role " + roleToAdd.Name + " to user: " + message.Author.Username)
	}
}

func removeAllRoles(session *discordgo.Session, member *discordgo.Member, rolesToRemove []string, guild *discordgo.Guild) {
	for _, roleToCheck := range rolesToRemove {
		if util.StrContains(member.Roles, roleToCheck, util.CaseSensitive) {
			session.GuildMemberRoleRemove(guild.ID, member.User.ID, roleToCheck)
		}
	}
}
