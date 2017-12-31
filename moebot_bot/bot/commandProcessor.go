package bot

import (
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
	"TEAM":      commTeam,
	"ROLE":      commRole,
	"RANK":      commRank,
	"NSFW":      commNsfw,
	"HELP":      commHelp,
	"CHANGELOG": commChange,
	"RAFFLE":    commRaffle,
	"SUBMIT":    commSubmit,
	"ECHO":      commEcho,
	"PERMIT":    commPermit,
	"CUSTOM":    commCustom,
	"PING":      commPing,
}

func RunCommand(session *discordgo.Session, message *discordgo.Message, guild *discordgo.Guild, channel *discordgo.Channel, member *discordgo.Member) {
	messageParts := strings.Split(message.Content, " ")
	if len(messageParts) <= 1 {
		// bad command, missing command after prefix
		return
	}
	command := strings.ToUpper(messageParts[1])

	if commFunc, commPresent := commands[command]; commPresent {
		log.Println("Processing command: " + command + " from user: " + fmt.Sprintf("%+v", message.Author))
		session.ChannelTyping(message.ChannelID)
		commFunc(&commPackage{session, message, guild, member, channel, messageParts[2:]})
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
	r := util.FindRole(pack.guild.Roles, roleName)
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
	pack.session.ChannelMessageSend(pack.channel.ID, "Added permission "+db.SprintPermission(permLevel)+" level for role "+roleName)
}

func commCustom(pack *commPackage) {
	// should have params: command name - role name
	if len(pack.params) < 2 {
		pack.session.ChannelMessageSend(pack.channel.ID, "Please provide command name followed by the role name")
		return
	}

	// get the role and server
	server, err := db.ServerQueryOrInsert(pack.guild.ID)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Error storing server information")
		return
	}
	roleName := strings.Join(pack.params[1:], " ")
	r := util.FindRole(pack.guild.Roles, roleName)
	role, err := db.RoleQueryOrInsert(db.Role{
		ServerId: server.Id,
		RoleUid:  r.ID,
	})

	err = db.CustomRoleAdd(pack.params[0], server.Id, role.Id)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Error adding custom role")
		return
	}
	pack.session.ChannelMessageSend(pack.channel.ID, "Added custom command `"+pack.params[0]+"` tied to the role: "+roleName)
}

func commEcho(pack *commPackage) {
	if m := checkValidMasterId(pack); m {
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
	pack.session.ChannelMessageSend(pack.channel.ID, "`"+ComPrefix+" role` has been renamed to `"+ComPrefix+" team`. Please use team instead!")
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

/*
Processes a guild role based on a list of allowed role names, and a requested role.
If StrictRole is true then the first role in the list of allowed roles is used if all roles are removed.
*/
func processGuildRole(allowedRoles []string, session *discordgo.Session, params []string, channel *discordgo.Channel, guild *discordgo.Guild, message *discordgo.Message, strictRole bool) {
	if len(params) == 0 || !util.StrContains(allowedRoles, params[0], util.CaseInsensitive) {
		session.ChannelMessageSend(channel.ID, message.Author.Mention()+" please provide one of the approved roles: "+strings.Join(allowedRoles, ", "))
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
		session.ChannelMessageSend(channel.ID, "You can't remove your lowest role for this command!")
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
