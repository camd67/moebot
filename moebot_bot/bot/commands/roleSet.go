package commands

import (
	"database/sql"
	"fmt"

	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type RoleSetCommand struct {
	ComPrefix string
}

func (rc *RoleSetCommand) Execute(pack *CommPackage) {
	args := ParseCommand(pack.params, []string{"-delete", "-role", "-trigger", "-confirm", "-security"})
	// should have params: command name - role name or delete - id
	if len(pack.params) < 2 {
		pack.session.ChannelMessageSend(pack.channel.ID, "Please provide command name followed by the role name")
		return
	}
	server, err := db.ServerQueryOrInsert(pack.guild.ID)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an error fetching this server. This is an error with moebot not discord!")
		return
	}
	if roleName, present := args["-delete"]; present {
		role := util.FindRoleByName(pack.guild.Roles, roleName)
		// we don't really care about the role itself here, just if we got a row back or not (could use a row count check but oh well)
		_, err := db.RoleQueryRoleUid(role.ID, server.Id)
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
	} else {
		roleName, hasRole := args["-role"]
		triggerName, hasTrigger := args["-trigger"]
		confirmText, hasConfirm := args["-confirm"]
		securityText, hasSecurity := args["-security"]
		if !hasRole && !hasTrigger && !hasConfirm && !hasSecurity {
			pack.session.ChannelMessageSend(pack.channel.ID, "This command requires both -role and another option provided (see help for details)")
			return
		}
		r := util.FindRoleByName(pack.guild.Roles, roleName)
		if r == nil {
			pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, it doesn't seem like that role exists on this server.")
			return
		}
		// first check if we've already got this one
		oldRole, err := db.RoleQueryRoleUid(r.ID, server.Id)
		var typeString string
		if err != nil {
			if err == sql.ErrNoRows {
				oldRole = db.Role{}
				typeString = "added"
				// don't return on a no row, that means we need to update
			} else {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an error finding that role. This is an error with moebot not discord!")
				return
			}
		} else {
			// we got a role back, so we're updating
			typeString = "updated"
		}
		if hasTrigger {
			if len(triggerName) < 0 || len(triggerName) > db.RoleMaxTriggerLength {
				pack.session.ChannelMessageSend(pack.channel.ID, "Please provide a trigger greater than 0 characters and less than "+
					db.RoleMaxTriggerLengthString+". The role was not updated.")
				return
			}
			oldRole.Trigger.Scan(triggerName)
		}
		if hasConfirm {
			if len(triggerName) < 0 || len(triggerName) > db.RoleMaxTriggerLength {
				pack.session.ChannelMessageSend(pack.channel.ID, "Please provide a trigger greater than 0 characters and less than "+
					db.RoleMaxTriggerLengthString+". The role was not updated.")
				return
			}
			oldRole.ConfirmationMessage.Scan(confirmText)
		}
		if hasSecurity {
			if len(securityText) < 0 || len(securityText) > db.RoleMaxTriggerLength {
				pack.session.ChannelMessageSend(pack.channel.ID, "Please provide a trigger greater than 0 characters and less than "+
					db.RoleMaxTriggerLengthString+". The role was not updated.")
				return
			}
			oldRole.ConfirmationSecurityAnswer.Scan(securityText)
		}
		err = db.RoleInsertOrUpdate(oldRole)
		if err != nil {
			pack.session.ChannelMessageSend(pack.channel.ID, "There was an error adding or updating the role. This is an issue with moebot and not discord")
			return
		}
		pack.session.ChannelMessageSend(pack.channel.ID, "Successfully "+typeString+" role information for "+roleName)
	}
}

func (rc *RoleSetCommand) GetPermLevel() db.Permission {
	return db.PermMod
}

func (rc *RoleSetCommand) GetCommandKeys() []string {
	return []string{"ROLESET"}
}

func (rc *RoleSetCommand) GetCommandHelp(commPrefix string) string {
	return fmt.Sprintf("`%[1]s roleset -roleName <role name> [-trigger <trigger> -confirm <confirmation message> -security <security code>]` - "+
		"Master/Mod. Provide roleName plus at least one other option.", commPrefix)
}
