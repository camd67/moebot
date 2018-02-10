package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type NsfwCommand struct{}

func (nc *NsfwCommand) Execute(pack *CommPackage) {
	// force NSFW comm param so we can reuse guild role
	processGuildRole([]string{"NSFW"}, pack.session, []string{"NSFW"}, pack.channel, pack.guild, pack.message, false)
}

func (nc *NsfwCommand) Setup(session *discordgo.Session) {}

func (nc *NsfwCommand) EventHandlers() []interface{} { return []interface{}{} }

func (nc *NsfwCommand) GetPermLevel() db.Permission {
	return db.PermAll
}
