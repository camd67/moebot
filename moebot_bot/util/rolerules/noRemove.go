package rolerules

import (
	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util/db/types"
)

type NoRemove struct {
	ReferenceGroup types.RoleGroup
}

func (r *NoRemove) Check(session *discordgo.Session, action *RoleAction) (success bool, message string) {
	if action.Action == RoleRemove {
		return false, "You've already got that role! You can change roles but can't remove them in the `" + r.ReferenceGroup.Name + "` group."
	}
	return true, ""
}

func (r *NoRemove) Apply(session *discordgo.Session, action *RoleAction) (success bool, message string) {
	return true, ""
}
