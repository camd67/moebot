package rolerules

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db/models"
	"github.com/camd67/moebot/moebot_bot/util/moeDiscord"
)

type Exclusive struct {
	ExclusiveRoles models.RoleSlice
}

func (r *Exclusive) Check(session *discordgo.Session, action *RoleAction) (success bool, message string) {
	return true, ""
}

func (r *Exclusive) Apply(session *discordgo.Session, action *RoleAction) (success bool, message string) {
	if action.Action == RoleRemove {
		return true, ""
	}
	var builder strings.Builder
	var err error
	foundOtherRole := false
	for _, dbGroupRole := range r.ExclusiveRoles {
		// only send a message that we removed the role if they actually have it and it's not the one we just added
		if dbGroupRole.RoleUID != action.Role.RoleUID && util.StrContains(action.Member.Roles, dbGroupRole.RoleUID, util.CaseSensitive) {
			roleToRemove := moeDiscord.FindRoleById(action.Guild.Roles, dbGroupRole.RoleUID)

			// Check for an error first
			err = session.GuildMemberRoleRemove(action.Guild.ID, action.Member.User.ID, dbGroupRole.RoleUID)
			if err != nil {
				builder.WriteString("\nFailed to remove: `")
				builder.WriteString(roleToRemove.Name)
				builder.WriteString("`")
				log.Println("error removing role from "+action.Member.User.ID+" with role "+action.Role.RoleUID+" with error: ", err)
				continue
			}

			// The user already has this role, remove it and tell them
			if !foundOtherRole {
				builder.WriteString("\nAlso removed:")
				foundOtherRole = true
			}
			builder.WriteString(" `")
			builder.WriteString(roleToRemove.Name)
			builder.WriteString("`")
		}
	}
	return err == nil, builder.String()
}
