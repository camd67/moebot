package bot

import (
	"log"
	"strconv"
	"strings"

	"github.com/camd67/moebot/moebot_bot/util"

	"github.com/bwmarrin/discordgo"
)

var commands = map[string]func(pack *commPackage){
	"TEAM": commTeam,
	// used for historical purposes since TEAM used to be ROLE
	"ROLE":      commRole,
	"RANK":      commRank,
	"NSFW":      commNsfw,
	"HELP":      commHelp,
	"CHANGELOG": commChange,
	"RAFFLE":    commRaffle,
	"SUBMIT":    commSubmit,
	"ECHO":      commEcho,
}

func RunCommand(session *discordgo.Session, message *discordgo.Message, guild *discordgo.Guild, channel *discordgo.Channel, member *discordgo.Member) {
	messageParts := strings.Split(message.Content, " ")
	if len(messageParts) <= 1 {
		// bad command, missing command after prefix
		return
	}
	command := strings.ToUpper(messageParts[1])

	if commFunc, commPresent := commands[command]; commPresent {
		session.ChannelTyping(message.ChannelID)
		commFunc(&commPackage{session, message, guild, member, channel, messageParts[2:]})
	}
}

func commEcho(pack *commPackage) {
	if pack.message.Author.ID != Config["masterId"] {
		pack.session.ChannelMessageSend(pack.message.ChannelID, "Sorry, only my master can use this command!")
		return
	}
	_, err := strconv.Atoi(pack.params[0])
	if err != nil {
		pack.session.ChannelMessageSend(pack.message.ChannelID, "Sorry, that's an invalid channel ID")
		return
	}
	pack.session.ChannelMessageSend(pack.params[0], strings.Join(pack.params[1:], " "))
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
