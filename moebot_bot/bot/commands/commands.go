package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type CommPackage struct {
	session *discordgo.Session
	message *discordgo.Message
	guild   *discordgo.Guild
	member  *discordgo.Member
	channel *discordgo.Channel
	params  []string
}

type Command interface {
	Execute(pack *CommPackage)
	Setup(session *discordgo.Session)
	EventHandlers() []interface{}
	GetPermLevel() db.Permission
}
