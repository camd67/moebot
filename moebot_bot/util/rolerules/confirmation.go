package rolerules

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db/types"
)

type Confirmation struct {
	ComPrefix string
}

func (r *Confirmation) Check(session *discordgo.Session, action *RoleAction) (success bool, message string) {
	if action.Action == RoleRemove {
		return true, ""
	}
	var confirmCodes []string
	for _, param := range strings.Split(action.OriginalMessage.Content, " ") {
		if strings.HasPrefix(param, "-") {
			confirmCodes = append(confirmCodes, param)
		}
	}
	// we only want to check for a confirmation when we have an actual confirmation message and they don't already have the role
	if action.Role.ConfirmationMessage.Valid && action.Role.ConfirmationMessage.String != "" && !util.StrContains(action.Member.Roles, action.Role.RoleUid, util.CaseSensitive) {
		// no confirm codes provided, given them their confirmation code
		if len(confirmCodes) <= 0 {
			err := r.sendConfirmationMessage(session, action.Channel, action.Role, action.Member)
			if err != nil {
				return false, "Sorry, I couldn't send you a PM! Please check your settings to allow direct messages from users on this server."
			}
			return true, action.Member.User.Mention() + " check your PM's for further instructions!"
		}

		session.ChannelMessageDelete(action.Channel.ID, action.OriginalMessage.ID)

		if action.Role.ConfirmationSecurityAnswer.Valid && action.Role.ConfirmationSecurityAnswer.String != "" {
			if len(confirmCodes) != 2 {
				return false, "Sorry, you need to insert a confirmation code and security answer to access " +
					"this role. Use `" + r.ComPrefix + " " + action.Role.Trigger.String + "` to receive a DM containing detailed instructions."
			}
			if !util.StrContains(confirmCodes, action.Role.ConfirmationSecurityAnswer.String, util.CaseSensitive) ||
				!util.StrContains(confirmCodes, "-"+r.getRoleCode(action.Role.RoleUid, action.Member.User.ID), util.CaseSensitive) {
				return false, "Sorry, you need to insert the correct confirmation code to access this role."
			}
		} else {
			if len(confirmCodes) != 1 {
				return false, "Sorry, you need to insert a confirmation code to access this role. Use `" +
					r.ComPrefix + " " + action.Role.Trigger.String + "` to receive a DM containing detailed instructions."
			}
			if !util.StrContains(confirmCodes, "-"+r.getRoleCode(action.Role.RoleUid, action.Member.User.ID), util.CaseSensitive) {
				return false, "Sorry, you need to insert the correct confirmation code to access this role."
			}
		}
	}
	return true, ""
}

func (r *Confirmation) Apply(session *discordgo.Session, action *RoleAction) (success bool, message string) {
	return true, ""
}

func (r *Confirmation) sendConfirmationMessage(session *discordgo.Session, channel *discordgo.Channel, role *types.Role, member *discordgo.Member) error {
	userChannel, err := session.UserChannelCreate(member.User.ID)
	if err != nil {
		// could log error creating user channel, but seems like it'll clutter the logs for a valid scenario..
		return err
	}
	roleCode := r.getRoleCode(role.RoleUid, member.User.ID)
	var messageText string
	if strings.Contains(strings.ToLower(role.ConfirmationMessage.String), types.RoleCodeSearchText) {
		messageText = strings.Replace(role.ConfirmationMessage.String, types.RoleCodeSearchText, roleCode, -1)
	} else {
		messageText = role.ConfirmationMessage.String + "\nYour confirmation code: `-" + roleCode + "`"
	}
	_, err = session.ChannelMessageSend(userChannel.ID, messageText)
	return err
}

/*
Returns a 6 character role code string that is unique per user and per role.
This should NOT be used for security features as it is not a secure algorithm.
*/
func (r *Confirmation) getRoleCode(roleUID, userUID string) string {
	hash := sha256.New()
	hash.Write([]byte(roleUID + userUID))
	return string(fmt.Sprintf("%x", hash.Sum(nil))[0:types.RoleCodeLength])
}
