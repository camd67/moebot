package commands

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type PermitCommand struct {
}

func (pc *PermitCommand) Execute(pack *CommPackage) {
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
	r := util.FindRoleByName(pack.guild.Roles, roleName)
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
	pack.session.ChannelMessageSend(pack.channel.ID, "Set permission ("+db.SprintPermission(permLevel)+") level for role "+roleName)
}

func (pc *PermitCommand) Setup(session *discordgo.Session) {}
func (pc *PermitCommand) EventHandlers() []interface{}     { return []interface{}{} }

func (pc *PermitCommand) GetPermLevel() db.Permission {
	return db.PermMaster
}
