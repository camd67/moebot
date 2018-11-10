package commands

import (
	"database/sql"
	"fmt"

	"github.com/camd67/moebot/moebot_bot/util/db"
	"github.com/camd67/moebot/moebot_bot/util/db/types"
	"github.com/camd67/moebot/moebot_bot/util/moeDiscord"
)

type PermitCommand struct {
}

func (pc *PermitCommand) Execute(pack *CommPackage) {
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
	r := moeDiscord.FindRoleByName(pack.guild.Roles, roleName)
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
			newGroupId, err := db.RoleGroupInsertOrUpdate(types.RoleGroup{
				ServerId: s.Id,
				Name:     db.UncategorizedGroup,
				Type:     types.GroupTypeAny,
			}, s)
			if err != nil {
				pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue adding an uncategorized group. This is an issue with moebot "+
					"not Discord")
				return
			}
			// Then update the returned dbRole to get the correct information
			dbRole.Groups = append(dbRole.Groups, newGroupId)
		} else {
			// if we got any other errors, then we want to bail out
			pack.session.ChannelMessageSend(pack.message.ChannelID, "Sorry, there was an issue retrieving that role. This is an issue with moebot and not discord")
			return
		}
	}
	// Overwrite the values even if present, since these should be the same
	dbRole.ServerId = s.Id
	dbRole.RoleUid = r.ID
	dbRole.Permission = permLevel
	err = db.RoleInsertOrUpdate(dbRole)
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, there was an issue editing that role. This is an issue with moebot not Discord.")
		return
	}
	pack.session.ChannelMessageSend(pack.channel.ID, "Edited role "+roleName+" successfully")
}

func (pc *PermitCommand) GetPermLevel() types.Permission {
	return types.PermGuildOwner
}

func (pc *PermitCommand) GetCommandKeys() []string {
	return []string{"PERMIT"}
}

func (pc *PermitCommand) GetCommandHelp(commPrefix string) string {
	return fmt.Sprintf("`%[1]s permit <role name> -permission <perm level>` - Edits the selected role to grant permission.", commPrefix)
}
