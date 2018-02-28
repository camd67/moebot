package commands

import (
	"crypto/sha256"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type RoleHandler struct {
	ComPrefix string
}

/*
Processes a guild role based on a list of allowed role names, and a requested role.
If StrictRole is true then the first role in the list of allowed roles is used if all roles are removed.
*/
func (r *RoleHandler) processGuildRole(allowedRoles []string, session *discordgo.Session, params []string, channel *discordgo.Channel, guild *discordgo.Guild, message *discordgo.Message, strictRole bool, sourceCommand string) {
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
	dbServer, err := db.ServerQueryOrInsert(guild.ID)
	if err != nil {
		log.Println("There was a problem querying or inserting the server", err)
		session.ChannelMessageSend(channel.ID, "Sorry, there was a problem retrieving server details, please try again.")
		return
	}
	dbRole, err := db.RoleQueryOrInsert(db.Role{
		ServerId: dbServer.Id,
		RoleUid:  roleToAdd.ID,
	})

	if dbRole.ConfirmationMessage.Valid && dbRole.ConfirmationMessage.String != "" && !util.StrContains(memberRoles, roleToAdd.ID, util.CaseSensitive) {
		if len(params) == 1 {
			err = r.sendConfirmationMessage(session, channel, dbRole, message.Author)
			if err == nil {
				session.ChannelMessageSend(channel.ID, message.Author.Mention()+" check your PM's for further instructions!")
			} else {
				session.ChannelMessageSend(channel.ID, "Sorry, I couldn't send you a PM! Please check your settings to allow direct messages from users on this server.")
			}
			return
		}
		session.ChannelMessageDelete(channel.ID, message.ID)
		if len(params) != 3 {
			session.ChannelMessageSend(channel.ID, "Sorry, you need to insert the correct confirmation code to access this role. Use `"+r.ComPrefix+" "+
				sourceCommand+"` to receive a DM containing detailed instructions.")
			return
		}
		if dbRole.ConfirmationSecurityAnswer.Valid && dbRole.ConfirmationSecurityAnswer.String != "" {
			if params[1] != dbRole.ConfirmationSecurityAnswer.String || params[2] != r.getRoleCode(roleToAdd.ID, message.Author.ID) {
				session.ChannelMessageSend(channel.ID, "Sorry, you need to insert the correct confirmation code to access this role.")
				return
			}
		} else {
			if params[1] != r.getRoleCode(roleToAdd.ID, message.Author.ID) {
				session.ChannelMessageSend(channel.ID, "Sorry, you need to insert the correct confirmation code to access this role.")
				return
			}
		}
	}

	r.removeAllRoles(session, guildMember, allRolesToChange, guild)
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

func (r *RoleHandler) removeAllRoles(session *discordgo.Session, member *discordgo.Member, rolesToRemove []string, guild *discordgo.Guild) {
	for _, roleToCheck := range rolesToRemove {
		if util.StrContains(member.Roles, roleToCheck, util.CaseSensitive) {
			session.GuildMemberRoleRemove(guild.ID, member.User.ID, roleToCheck)
		}
	}
}

func (r *RoleHandler) sendConfirmationMessage(session *discordgo.Session, channel *discordgo.Channel, role db.Role, user *discordgo.User) error {
	userChannel, err := session.UserChannelCreate(user.ID)
	if err != nil {
		// could log error creating user channel, but seems like it'll clutter the logs for a valid scenario..
		return err
	}
	_, err = session.ChannelMessageSend(userChannel.ID, fmt.Sprintf(role.ConfirmationMessage.String, r.getRoleCode(role.RoleUid, user.ID)))
	return err
}

func (r *RoleHandler) getRoleCode(roleUID, userUID string) string {
	hash := sha256.New()
	hash.Write([]byte(roleUID + userUID))
	return string(fmt.Sprintf("%x", hash.Sum(nil))[0:6])
}
