package commands

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/bot/permissions"
	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type RoleCommand struct {
	ComPrefix   string
	PermChecker permissions.PermissionChecker
}

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
		printAllRoles(server, vetRole, pack)
	} else {
		var role *discordgo.Role
		var roleGroup db.RoleGroup
		var dbRole db.Role
		var confirmCodes []string
		// TODO: convert this to use the actual veteran name
		if strings.EqualFold(pack.params[0], "veteran") {
			// before anything, if the server doesn't have a rank or role bail out
			if !server.VeteranRank.Valid || !server.VeteranRole.Valid {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, this server isn't setup to handle veteran role yet! Contact the server admins.")
				return
			}

			// make some placeholder veteran role tables
			roleGroup = db.RoleGroup{
				Name: "Veteran",
				Type: db.GroupTypeExclusive,
			}
			dbRole = db.Role{
				RoleUid: server.VeteranRole.String,
				Trigger: sql.NullString{
					String: "veteran",
					Valid:  true,
				},
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
			// load up the trigger to see if it exists, stripping out anything prefixed with - (our security text)
			var roleNameBuf bytes.Buffer
			for _, param := range pack.params {
				if !strings.HasPrefix(param, "-") {
					roleNameBuf.WriteString(param)
					roleNameBuf.WriteString(" ")
				} else {
					confirmCodes = append(confirmCodes, param)
				}
			}
			dbRole, err = db.RoleQueryTrigger(strings.Trim(roleNameBuf.String(), " "), server.Id)
			// an invalid trigger should pretty much never happen, but checking for it anyways
			if err != nil || !dbRole.Trigger.Valid {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue fetching the role. Please provide a valid role. `"+
					rc.ComPrefix+" role` to list all roles for this server.")
				return
			}
			role = util.FindRoleById(pack.guild.Roles, dbRole.RoleUid)
			if role == nil {
				log.Println("Nil dbRole when searching for dbRole id:" + dbRole.RoleUid)
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue finding that role in this server. It may have been deleted.")
				return
			}
			roleGroup, err = db.RoleGroupQueryId(dbRole.GroupId)
			if err != nil {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, There was an issue finding that role group. This is an issue with moebot "+
					"and not discord.")
				return
			}
		}
		// process the role to see if it has a confirmation message, then decide if we need to bail out or continue to the role update phase
		if !rc.processRoleConfirmation(dbRole, role, pack, confirmCodes) {
			return
		}

		rc.updateUserRoles(pack, role, roleGroup)
	}
}

/*
Actually go through and update the roles for this user based on the given role and role group
*/
func (rc *RoleCommand) updateUserRoles(pack *CommPackage, role *discordgo.Role, group db.RoleGroup) {
	if util.StrContains(pack.member.Roles, role.ID, util.CaseSensitive) {
		if group.Type == db.GroupTypeExclusiveNoRemove {
			pack.session.ChannelMessageSend(pack.channel.ID, "You've already got that role! You can change roles but can't remove them in the `"+
				group.Name+"` group.")
		} else {
			pack.session.GuildMemberRoleRemove(pack.guild.ID, pack.message.Author.ID, role.ID)
			pack.session.ChannelMessageSend(pack.channel.ID, "Removed role "+role.Name+" for "+pack.message.Author.Mention())
		}
	} else {
		if group.Type == db.GroupTypeAny {
			pack.session.GuildMemberRoleAdd(pack.guild.ID, pack.message.Author.ID, role.ID)
			pack.session.ChannelMessageSend(pack.channel.ID, "Added role "+role.Name+" for "+pack.message.Author.Mention())
		} else {
			// This case needs to check to see if the user has any other roles from this group, since they may not be allowed to add more
			// (GroupTypeExclusive, GroupTypeExclusiveNoRemove)
			fullGroupRoles, err := db.RoleQueryGroup(group.Id)
			if err != nil {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, There was an issue finding that role group. This is an issue with moebot "+
					"and not discord.")
				return
			}
			// we'll always be adding a role here
			pack.session.GuildMemberRoleAdd(pack.guild.ID, pack.message.Author.ID, role.ID)
			var message bytes.Buffer
			message.WriteString("Added role `")
			message.WriteString(role.Name)
			message.WriteString("` for ")
			message.WriteString(pack.message.Author.Mention())
			// we should only find one other role, but just in case
			foundOtherRole := false
			for _, dbGroupRole := range fullGroupRoles {
				if util.StrContains(pack.member.Roles, dbGroupRole.RoleUid, util.CaseSensitive) {
					roleToRemove := util.FindRoleById(pack.guild.Roles, dbGroupRole.RoleUid)
					// The user already has this role, remove it and tell them
					if !foundOtherRole {
						message.WriteString("\nAlso removed:")
						foundOtherRole = true
					}
					message.WriteString(" `")
					message.WriteString(roleToRemove.Name)
					message.WriteString("`")
					pack.session.GuildMemberRoleRemove(pack.guild.ID, pack.message.Author.ID, dbGroupRole.RoleUid)
				}
			}
			pack.session.ChannelMessageSend(pack.channel.ID, message.String())
		}
	}
}

func (rc *RoleCommand) processRoleConfirmation(dbRole db.Role, roleToAdd *discordgo.Role, pack *CommPackage, confirmCodes []string) (shouldProceed bool) {
	// we only want to check for a confirmation when we have an actual confirmation message and they don't already have the role
	if dbRole.ConfirmationMessage.Valid && dbRole.ConfirmationMessage.String != "" && !util.StrContains(pack.member.Roles, roleToAdd.ID, util.CaseSensitive) {
		// no confirm codes provided, given them their confirmation code
		if len(confirmCodes) <= 0 {
			err := rc.sendConfirmationMessage(pack.session, pack.channel, dbRole, pack.message.Author)
			if err == nil {
				pack.session.ChannelMessageSend(pack.channel.ID, pack.message.Author.Mention()+" check your PM's for further instructions!")
			} else {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, I couldn't send you a PM! Please check your settings to allow direct messages from "+
					"users on this server.")
			}
			return false
		}

		// they gave us some sort of confirmation code, delete their message
		pack.session.ChannelMessageDelete(pack.channel.ID, pack.message.ID)

		// Optionally check for a security answer, since we can have just a confirmation code and no security
		if dbRole.ConfirmationSecurityAnswer.Valid && dbRole.ConfirmationSecurityAnswer.String != "" {
			if len(confirmCodes) != 2 {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, you need to insert a confirmation code and security answer to access "+
					"this role. Use `"+rc.ComPrefix+" "+dbRole.Trigger.String+"` to receive a DM containing detailed instructions.")
				return false
			}
			if !util.StrContains(confirmCodes, dbRole.ConfirmationSecurityAnswer.String, util.CaseSensitive) ||
				!util.StrContains(confirmCodes, "-"+rc.getRoleCode(roleToAdd.ID, pack.message.Author.ID), util.CaseSensitive) {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, you need to insert the correct confirmation code to access this role.")
				return false
			}
		} else {
			if len(confirmCodes) != 1 {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, you need to insert a confirmation code to access this role. Use `"+
					rc.ComPrefix+" "+dbRole.Trigger.String+"` to receive a DM containing detailed instructions.")
				return false
			}
			if !util.StrContains(confirmCodes, "-"+rc.getRoleCode(roleToAdd.ID, pack.message.Author.ID), util.CaseSensitive) {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, you need to insert the correct confirmation code to access this role.")
				return false
			}
		}
	}
	return true
}
func (rc *RoleCommand) sendConfirmationMessage(session *discordgo.Session, channel *discordgo.Channel, role db.Role, user *discordgo.User) error {
	userChannel, err := session.UserChannelCreate(user.ID)
	if err != nil {
		// could log error creating user channel, but seems like it'll clutter the logs for a valid scenario..
		return err
	}
	message := role.ConfirmationMessage.String + "\nYour confirmation code: `-" + rc.getRoleCode(role.RoleUid, user.ID) + "`"
	_, err = session.ChannelMessageSend(userChannel.ID, message)
	return err
}

func (rc *RoleCommand) getRoleCode(roleUID, userUID string) string {
	hash := sha256.New()
	hash.Write([]byte(roleUID + userUID))
	return string(fmt.Sprintf("%x", hash.Sum(nil))[0:6])
}

func (rc *RoleCommand) GetPermLevel() db.Permission {
	return db.PermAll
}

func (rc *RoleCommand) GetCommandKeys() []string {
	return []string{"ROLE"}
}

func (rc *RoleCommand) GetCommandHelp(commPrefix string) string {
	return fmt.Sprintf("`%[1]s role <role name>` - Changes your role to one of the approved roles. `%[1]s role` to list all the roles", commPrefix)
}
func printAllRoles(server db.Server, vetRole *discordgo.Role, pack *CommPackage) {
	triggersByGroup := make(map[string][]string)
	// go find all the roles for this server
	roles, err := db.RoleQueryServer(server)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue fetching the server. This is an issue with moebot!")
		return
	}
	// Then find all the groups for the server
	roleGroups, err := db.RoleGroupQueryServer(server)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue fetching the roles for this server. This is an issue with moebot!")
		return
	}
	if vetRole != nil {
		// triggers = append(triggers, vetRole.Name)
		// TODO: this should be the name of the role, but role is restricted to one word right now...
		triggersByGroup["veteran"] = append(triggersByGroup["veteran"], "veteran")
	}
	for _, role := range roles {
		if !role.Trigger.Valid {
			// skip any invalid triggers. We don't want people thinking that they can choose roles they actually can't
			continue
		}
		// Could maybe make a map here, but the group size is going to be pretty small
		foundGroup := false
		for _, group := range roleGroups {
			if role.GroupId == group.Id {
				triggersByGroup[group.Name] = append(triggersByGroup[group.Name], role.Trigger.String)
				foundGroup = true
			}
		}
		if !foundGroup {
			log.Println("!!! WARNING !!! Failed to find group for a role! This is most likely a programming error")
			triggersByGroup["uncategorized"] = append(triggersByGroup["uncategorized"], role.Trigger.String)
		}
	}
	var message bytes.Buffer
	if len(triggersByGroup) == 0 {
		message.WriteString("Looks like there aren't any roles I can assign to you in this server!")
	} else {
		message.WriteString("This server's roles (highlighted `like this`): ")
		for groupName, triggerList := range triggersByGroup {
			message.WriteString("\nGroup (")
			message.WriteString(groupName)
			// TODO: add group type string here
			message.WriteString("): `")
			message.WriteString(strings.Join(triggerList, "`, `"))
			message.WriteString("`. ")
		}
	}
	pack.session.ChannelMessageSend(pack.channel.ID, message.String())
}
