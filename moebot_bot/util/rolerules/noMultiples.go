package rolerules

import (
	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db/models"
	"github.com/camd67/moebot/moebot_bot/util/moeDiscord"
)

type NoMultiples struct {
	ExclusiveRoles models.RoleSlice
}

func (r *NoMultiples) Check(session *discordgo.Session, action *RoleAction) (success bool, message string) {
	if action.Action == RoleRemove {
		return true, ""
	}
	for _, dbGroupRole := range r.ExclusiveRoles {
		// only send a message that we removed the role if they actually have it and it's not the one we just added
		if dbGroupRole.RoleUID != action.Role.RoleUID && util.StrContains(action.Member.Roles, dbGroupRole.RoleUID, util.CaseSensitive) {
			existingRole := moeDiscord.FindRoleById(action.Guild.Roles, dbGroupRole.RoleUID)
			return false, "You already have the role `" + existingRole.Name + "`, please remove it before adding this role."
		}
	}
	return true, ""
}

func (r *NoMultiples) Apply(session *discordgo.Session, action *RoleAction) (success bool, message string) {
	return true, ""
}
