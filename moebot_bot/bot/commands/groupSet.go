package commands

import (
	"bytes"
	"database/sql"
	"fmt"

	"github.com/camd67/moebot/moebot_bot/util/db"
)

type GroupSetCommand struct {
	ComPrefix string
}

func (gc *GroupSetCommand) Execute(pack *CommPackage) {
	args := ParseCommand(pack.params, []string{"-delete", "-name", "-type"})
	deleteName, hasDelete := args["-delete"]
	groupName, hasName := args["-name"]
	typeText, hasType := args["-type"]

	server, err := db.ServerQueryOrInsert(pack.guild.ID)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an error fetching this server. This is an error with moebot not discord!")
		return
	}

	if !hasDelete || (!hasName && !hasType) {
		// error state, they didn't give anything
		groups, err := db.RoleGroupQueryServer(server)
		if err != nil {
			pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an error fetching groups for this server. This is an error with moebot "+
				"not discord!")
			return
		}
		if len(groups) <= 0 {
			pack.session.ChannelMessageSend(pack.channel.ID, "It doesn't look like there's any groups in this server! You can make them with this "+
				"command though")
			return
		}
		// print all the group names
		var message bytes.Buffer
		message.WriteString("Groups in this server: ")
		for _, g := range groups {
			message.WriteString("`")
			message.WriteString(g.Name)
			message.WriteString("`-Type(`")
			message.WriteString(db.GetStringFromGroupType(g.Type))
			message.WriteString("`), ")
		}
		pack.session.ChannelMessageSend(pack.channel.ID, message.String())
	} else if hasDelete {
		// we want to delete the group they gave us (if it exists)
		dbRoleGroup, err := db.RoleGroupQueryName(deleteName, server.Id)
		if err != nil {
			if err == sql.ErrNoRows {
				pack.session.ChannelMessageSend(pack.channel.ID, "It doesn't look like that's a group you can delete! Please provide a group that was "+
					"previously set up")
			} else {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an error finding that group. This is an error with moebot not discord!")
			}
			return
		}
		err = db.RoleGroupDelete(dbRoleGroup.Id)
		if err != nil {
			pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an error deleting that role. This is an error with moebot not discord!")
			return
		}
		pack.session.ChannelMessageSend(pack.channel.ID, "Deleted "+deleteName+"!")
	} else {
		// add in a new group, or update an existing one
		dbRoleGroup, err := db.RoleGroupQueryName(groupName, server.Id)
		var dbOperationType string
		if err != nil {
			if err == sql.ErrNoRows {
				dbOperationType = "added"
				dbRoleGroup = db.RoleGroup{}
			} else {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an error finding that role group. This is an error with moebot "+
					"not discord!")
				return
			}
		} else {
			dbOperationType = "updated"
		}
		if len(groupName) < 0 || len(groupName) > db.RoleGroupMaxNameLength {
			pack.session.ChannelMessageSend(pack.channel.ID, "You must provide a valid group name, less than "+db.RoleGroupMaxNameLengthString+
				" characters long")
			return
		}
		dbRoleGroup.Name = groupName
		newType := db.GetGroupTypeFromString(typeText)
		if newType < 0 {
			// invalid type
			pack.session.ChannelMessageSend(pack.channel.ID, "You must provide a valid group type from the following: "+db.OptionsForGroupType)
			return
		}
		dbRoleGroup.Type = newType
		err = db.RoleGroupInsertOrUpdate(dbRoleGroup, server)
		if err != nil {
			pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue updating the role group. This most likely means your change "+
				"wasn't applied")
			return
		}
		pack.session.ChannelMessageSend(pack.channel.ID, "Successfully "+dbOperationType+" the group "+groupName)
	}
}

func (gc *GroupSetCommand) GetPermLevel() db.Permission {
	return db.PermMod
}

func (gc *GroupSetCommand) GetCommandKeys() []string {
	return []string{"GROUPSET"}
}

func (gc *GroupSetCommand) GetCommandHelp(commPrefix string) string {
	return fmt.Sprintf("`%[1]s groupset -name <group name> -type <type>` - Master/Mod. Creates a new group with the given name and type for this server."+
		" Use `-delete <group name>` to delete a group.", commPrefix)
}
