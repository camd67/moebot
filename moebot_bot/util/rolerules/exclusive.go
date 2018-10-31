package rolerules

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db/types"
	"github.com/camd67/moebot/moebot_bot/util/moeDiscord"
)

type Exclusive struct {
	ExclusiveRoles []types.Role
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
		if dbGroupRole.RoleUid != action.Role.RoleUid && util.StrContains(action.Member.Roles, dbGroupRole.RoleUid, util.CaseSensitive) {
			roleToRemove := moeDiscord.FindRoleById(action.Guild.Roles, dbGroupRole.RoleUid)

			// Check for an error first
			err = session.GuildMemberRoleRemove(action.Guild.ID, action.Member.User.ID, dbGroupRole.RoleUid)
			if err != nil {
				builder.WriteString("\nFailed to remove: `")
				builder.WriteString(roleToRemove.Name)
				builder.WriteString("`")
				log.Println("error removing role from "+action.Member.User.ID+" with role "+action.Role.RoleUid+" with error: ", err)
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
	return err != nil, builder.String()
}
