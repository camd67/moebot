package rolerules

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type Points struct {
	PointsTreshold int
}

func (r *Points) Check(session *discordgo.Session, action *RoleAction) (success bool, message string) {
	if action.Action == RoleRemove {
		return true, ""
	}
	if action.UserRank == nil {
		return false, "Sorry, you don't have enough veteran points yet! You're currently: Unranked"
	}
	pointCountMessage := fmt.Sprintf("%.2f%% of the way to veteran", float64(action.UserRank.Rank)/float64(r.PointsTreshold)*100)
	if action.UserRank.Rank < r.PointsTreshold {
		return false, "Sorry, you don't have enough veteran points yet! You're currently: " + pointCountMessage
	}
	return true, ""
}

func (r *Points) Apply(session *discordgo.Session, action *RoleAction) (success bool, message string) {
	return true, ""
}
