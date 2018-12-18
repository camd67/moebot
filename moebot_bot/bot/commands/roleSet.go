package commands

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"
	"github.com/camd67/moebot/moebot_bot/util/db/models"
	"github.com/camd67/moebot/moebot_bot/util/db/types"
	"github.com/camd67/moebot/moebot_bot/util/moeDiscord"
)

type RoleSetCommand struct {
	ComPrefix string
}

func (rc *RoleSetCommand) Execute(pack *CommPackage) {
	args := ParseCommand(pack.params, []string{"-delete", "-role", "-trigger", "-confirm", "-security", "-group"})

	server, err := db.ServerQueryOrInsert(pack.guild.ID)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an error fetching this server. This is an error with moebot not discord!")
		return
	}
	deleteName, hasDelete := args["-delete"]
	roleName, hasRole := args["-role"]
	triggerName, hasTrigger := args["-trigger"]
	confirmText, hasConfirm := args["-confirm"]
	securityText, hasSecurity := args["-security"]
	groupText, hasGroup := args["-group"]

	if !hasDelete && !hasRole && !hasTrigger && !hasConfirm && !hasSecurity && !hasGroup {
		// empty command (or just really bad one)
		var vetRole *discordgo.Role
		if server.VeteranRole.Valid {
			vetRole = moeDiscord.FindRoleById(pack.guild.Roles, server.VeteranRole.String)
		}
		if len(pack.params) == 0 {
			printAllRoles(server, vetRole, pack)
		}
	} else if hasDelete {
		rc.deleteRole(deleteName, pack, server)
		return
	} else {
		if !hasRole {
			pack.session.ChannelMessageSend(pack.channel.ID, "This command requires a role (supplied with -role)")
			return
		}
		if !hasTrigger && !hasConfirm && !hasSecurity && !hasGroup {
			pack.session.ChannelMessageSend(pack.channel.ID, "You must provide at least one of: trigger, confirm, group, or security")
			return
		}

		r := moeDiscord.FindRoleByName(pack.guild.Roles, roleName)
		if r == nil {
			pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, it doesn't seem like that role exists on this server.")
			return
		}
		// first check if we've already got this one
		oldRole, err := db.RoleQueryRoleUid(r.ID, server.ID)
		var oldGroups models.RoleGroupSlice
		var typeString string
		if err != nil {
			if err == sql.ErrNoRows {
				oldRole = &models.Role{
					RoleUID: r.ID,
				}
				typeString = "added"
				// don't return on a no row error, that means we need to add a new role
				// validate to make sure we got the required information for a new role as opposed to an update
				if !hasGroup || !hasTrigger {
					pack.session.ChannelMessageSend(pack.channel.ID, "You must provide a group and trigger when making new roles")
					return
				}
			} else {
				// we got an actual error
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an error finding that role. This is an error with moebot not discord!")
				return
			}
		} else {
			// we got a role back, so we're updating
			oldGroups = oldRole.R.RoleGroups
			typeString = "updated"
		}
		if hasTrigger {
			if len(triggerName) < 0 || len(triggerName) > db.RoleMaxTriggerLength {
				pack.session.ChannelMessageSend(pack.channel.ID, "Please provide a trigger greater than 0 characters and less than "+
					db.RoleMaxTriggerLengthString+". The role was not updated.")
				return
			}
			oldRole.Trigger.Scan(strings.TrimSpace(triggerName))
		}
		if hasConfirm {
			if len(confirmText) < 0 || len(confirmText) > db.MaxMessageLength {
				pack.session.ChannelMessageSend(pack.channel.ID, "Please provide a confirmation text greater than 0 characters and less than "+
					db.MaxMessageLengthString+". The role was not updated.")
				return
			}

			// Make a dummy string that is as long as a role code for testing the length, plus one more for the hyphen
			exampleRoleCode := strings.Repeat("#", types.RoleCodeLength+1)
			if strings.Contains(strings.ToLower(confirmText), types.RoleCodeSearchText) {
				postReplacementLength := len(strings.Replace(confirmText, types.RoleCodeSearchText, exampleRoleCode, -1))
				if postReplacementLength > moeDiscord.MaxMessageLength {
					charactersOver := postReplacementLength - moeDiscord.MaxMessageLength
					pack.session.ChannelMessageSend(pack.channel.ID, "When replacing every instance of "+types.RoleCodeSearchText+" with a "+
						strconv.Itoa(types.RoleCodeLength+1)+" character role code your confirmation message went over the discord max message length. "+
						"Please shorten it by "+strconv.Itoa(charactersOver))
					return
				}
			}
			oldRole.ConfirmationMessage.Scan(confirmText)
		}
		if hasSecurity {
			if len(securityText) < 0 || len(securityText) > db.MaxMessageLength {
				pack.session.ChannelMessageSend(pack.channel.ID, "Please provide a security text greater than 0 characters and less than "+
					db.MaxMessageLengthString+". The role was not updated.")
				return
			}
			if !strings.HasPrefix(securityText, "-") {
				// append on a - to the front of the security code
				securityText = "-" + securityText
			}
			oldRole.ConfirmationSecurityAnswer.Scan(securityText)
		}

		group, err := db.RoleGroupQueryName(groupText, server.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				pack.session.ChannelMessageSend(pack.channel.ID, "You must provide a group that exists. You can create this with the groupset command.")
			} else {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue querying for the provided group. This is an issue with moebot "+
					"and not discord.")
			}
			return
		}
		groups, err := updateRoleGroups(server, oldGroups, group)
		if err != nil {
			pack.session.ChannelMessageSend(pack.channel.ID, "There was an error updating role groups. This is an issue with moebot and not discord")
			return
		}

		oldRole.ServerID.Int = server.ID
		err = db.RoleInsertOrUpdate(oldRole, groups)
		if err != nil {
			pack.session.ChannelMessageSend(pack.channel.ID, "There was an error adding or updating the role. This is an issue with moebot and not discord")
			return
		}
		pack.session.ChannelMessageSend(pack.channel.ID, "Successfully "+typeString+" role information for "+roleName)
	}
}

func updateRoleGroups(server *models.Server, groups models.RoleGroupSlice, group *models.RoleGroup) (models.RoleGroupSlice, error) {
	defaultGroup, err := db.RoleGroupQueryName(db.UncategorizedGroup, server.ID)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error while retrieving the default role group", err)
		return nil, err
	}
	groups = util.GroupRemove(groups, defaultGroup.ID)
	if util.GroupContains(groups, group.ID) {
		groups = util.GroupRemove(groups, group.ID)
		if len(groups) == 0 && defaultGroup.ID != 0 {
			groups = append(groups, &models.RoleGroup{ID: defaultGroup.ID})
		}
	} else {
		groups = append(groups, &models.RoleGroup{ID: group.ID})
	}
	return groups, nil
}

func (rc *RoleSetCommand) GetPermLevel() types.Permission {
	return types.PermMod
}

func (rc *RoleSetCommand) GetCommandKeys() []string {
	return []string{"ROLESET"}
}

func (rc *RoleSetCommand) GetCommandHelp(commPrefix string) string {
	return fmt.Sprintf("`%[1]s roleset -role <role name> [-trigger <trigger> -confirm <confirmation message> -security <security code> "+
		"-group <group name>]` - Master/Mod. Provide roleName plus at least one other option. Security code must be prefixed with `-` in your "+
		"confirmation message if you want to include it.", commPrefix)
}

func (rc *RoleSetCommand) deleteRole(roleName string, pack *CommPackage, server *models.Server) {
	role := moeDiscord.FindRoleByName(pack.guild.Roles, roleName)
	// we don't really care about the role itself here, just if we got a row back or not (could use a row count check but oh well)
	_, err := db.RoleQueryRoleUid(role.ID, server.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			pack.session.ChannelMessageSend(pack.channel.ID, "It doesn't look like that's a role you can delete! Please provide a role that was "+
				"previously set up")
		} else {
			pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an error finding that role. This is an error with moebot not discord!")
		}
		return
	}
	err = db.RoleDelete(role.ID, pack.guild.ID)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an error deleting that role. This is an error with moebot not discord!")
		return
	}
	pack.session.ChannelMessageSend(pack.channel.ID, "Deleted "+roleName+"!")
}
