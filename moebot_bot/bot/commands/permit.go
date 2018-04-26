package commands

import (
	"database/sql"
	"fmt"

	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type PermitCommand struct {
}

func (pc *PermitCommand) Execute(pack *CommPackage) {
	// todo: Could probably migrate this over to the role command. Keeping security answer and confirmation in here for now though
	args := ParseCommand(pack.params, []string{"-permission"})
	roleName, rolePresent := args[""]
	permText, permPresent := args["-permission"]

	if !rolePresent || !permPresent {
		pack.session.ChannelMessageSend(pack.message.ChannelID, "Please provide a role name followed by -permission and a permission level.")
		return
	}

	permLevel := db.GetPermissionFromString(permText)
	if !db.IsAssignablePermissionLevel(permLevel) {
		pack.session.ChannelMessageSend(pack.message.ChannelID, "Invalid permission level. Valid levels: "+db.GetAssignableRoles())
		return
	}

	// find the correct role
	r := util.FindRoleByName(pack.guild.Roles, roleName)
	if r == nil {
		pack.session.ChannelMessageSend(pack.message.ChannelID, "Please provide a role that exists in this server")
	}
	// we've got the role, add it to the db, updating if necessary
	// but first grab the server (probably want to move this out to include in the commPackage
	s, err := db.ServerQueryOrInsert(pack.guild.ID)
	if err != nil {
		pack.session.ChannelMessageSend(pack.message.ChannelID, "Error retrieving server information. This is an issue with moebot and not Discord")
		return
	}
	// Then check to see if the role exists in the server
	dbRole, err := db.RoleQueryRoleUid(r.ID, s.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			// we don't want to return on a no row error, instead add a default group so we can add later
			err = db.RoleGroupInsertOrUpdate(db.RoleGroup{
				ServerId: s.Id,
				Name:     db.UncategorizedGroup,
				Type:     db.GroupTypeAny,
			}, s)
			if err != nil {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue adding an uncategorized group. This is an issue with moebot "+
					"not Discord")
				return
			}
		} else {
			// if we got any other errors, then we want to bail out
			pack.session.ChannelMessageSend(pack.message.ChannelID, "Sorry, there was an issue retrieving that role. This is an issue with moebot and not discord")
			return
		}
	}
	dbRole.Permission = permLevel
	db.RoleInsertOrUpdate(dbRole)
	pack.session.ChannelMessageSend(pack.channel.ID, "Edited role "+roleName+" successfully")
}

func (pc *PermitCommand) GetPermLevel() db.Permission {
	return db.PermGuildOwner
}

func (pc *PermitCommand) GetCommandKeys() []string {
	return []string{"PERMIT"}
}

func (pc *PermitCommand) GetCommandHelp(commPrefix string) string {
	return fmt.Sprintf("`%[1]s permit <role name> -permission <perm level>` - Edits the selected role to grant permission.", commPrefix)
}
