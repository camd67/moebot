package rolerules

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util/db"
	"github.com/camd67/moebot/moebot_bot/util/db/models"
	"github.com/camd67/moebot/moebot_bot/util/db/types"
)

//RoleActionType Type of action being performed on the role
type RoleActionType int

//RoleAction Action being performed on the role
type RoleAction struct {
	Role            *models.Role
	UserRank        *models.UserServerRank
	Member          *discordgo.Member
	Guild           *discordgo.Guild
	Channel         *discordgo.Channel
	OriginalMessage *discordgo.Message
	Action          RoleActionType
}

const (
	//RoleAdd Adding the role
	RoleAdd RoleActionType = 1
	//RoleRemove Removing the role
	RoleRemove RoleActionType = 2
)

//RoleRule Rule defining if and how a role should be applied
type RoleRule interface {
	Check(session *discordgo.Session, action *RoleAction) (success bool, message string)
	Apply(session *discordgo.Session, action *RoleAction) (success bool, message string)
}

func GetRulesForRole(server *models.Server, role *models.Role, comPrefix string) ([]RoleRule, error) {
	var result []RoleRule
	if server.VeteranRole.String == role.RoleUID && server.VeteranRank.Valid {
		result = append(result, &Points{PointsTreshold: int(server.VeteranRank.Int)})
	}
	if role.ConfirmationMessage.Valid {
		result = append(result, &Confirmation{ComPrefix: comPrefix})
	}
	for _, g := range role.R.RoleGroups {
		group, err := db.RoleGroupQueryId(g.ID)
		if err != nil {
			log.Println("Error while retrieving role group during rules initialization", err)
			return nil, err
		}
		relatedRoles, err := db.RoleQueryGroup(g.ID)
		if err != nil {
			log.Println("Error while retrieving related group roles during rules initialization", err)
			return nil, err
		}
		switch group.GroupType {
		case types.GroupTypeExclusive:
			result = append(result, &Exclusive{ExclusiveRoles: relatedRoles})
			break
		case types.GroupTypeExclusiveNoRemove:
			result = append(result, &Exclusive{ExclusiveRoles: relatedRoles})
			result = append(result, &NoRemove{ReferenceGroup: group})
			break
		case types.GroupTypeNoMultiples:
			result = append(result, &NoMultiples{ExclusiveRoles: relatedRoles})
			break
		}
	}
	return result, nil
}
