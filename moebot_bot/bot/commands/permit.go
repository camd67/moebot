package commands

import (
	"strings"

	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type PermitCommand struct {
	checker PermissionChecker
}

func (pc PermitCommand) Execute(pack *CommPackage) {
	if m := pc.checker.hasValidMasterId(pack); !m {
		return
	}
	// should always have more than 2 params: permission level, role name ... role name
	if len(pack.Params) < 2 {
		pack.Session.ChannelMessageSend(pack.Message.ChannelID, "Please provide a permission level followed by the role name")
		return
	}
	permLevel := db.GetPermissionFromString(pack.Params[0])
	if permLevel == -1 {
		pack.Session.ChannelMessageSend(pack.Message.ChannelID, "Invalid permission level")
		return
	}
	// find the correct role
	roleName := strings.Join(pack.Params[1:], " ")
	r := util.FindRoleByName(pack.Guild.Roles, roleName)
	if r == nil {
		pack.Session.ChannelMessageSend(pack.Message.ChannelID, "Unknown role name")
	}
	// we've got the role, add it to the db, updating if necessary
	// but first grab the server (probably want to move this out to include in the commPackage
	s, err := db.ServerQueryOrInsert(pack.Guild.ID)
	if err != nil {
		pack.Session.ChannelMessageSend(pack.Message.ChannelID, "Error retrieving server information")
		return
	}
	db.RoleInsertOrUpdate(db.Role{
		ServerId:   s.Id,
		RoleUid:    r.ID,
		Permission: permLevel,
	})
	pack.Session.ChannelMessageSend(pack.Channel.ID, "Set permission ("+db.SprintPermission(permLevel)+") level for role "+roleName)
}
