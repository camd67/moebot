package commands

import (
	"crypto/sha256"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util/db/models"
)

type RoleHandler struct {
	ComPrefix string
}

func (r *RoleHandler) sendConfirmationMessage(session *discordgo.Session, channel *discordgo.Channel, role models.Role, user *discordgo.User) error {
	userChannel, err := session.UserChannelCreate(user.ID)
	if err != nil {
		// could log error creating user channel, but seems like it'll clutter the logs for a valid scenario..
		return err
	}
	_, err = session.ChannelMessageSend(userChannel.ID, fmt.Sprintf(role.ConfirmationMessage.String, r.getRoleCode(role.RoleUID, user.ID)))
	return err
}

func (r *RoleHandler) getRoleCode(roleUID, userUID string) string {
	hash := sha256.New()
	hash.Write([]byte(roleUID + userUID))
	return string(fmt.Sprintf("%x", hash.Sum(nil))[0:6])
}
