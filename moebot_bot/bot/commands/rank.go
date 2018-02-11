package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type RankCommand struct{}

func (rc *RankCommand) Execute(pack *CommPackage) {
	processGuildRole([]string{"Red", "Blue"}, pack.session, pack.params, pack.channel, pack.guild, pack.message, true)
}

func (rc *RankCommand) Setup(session *discordgo.Session) {}

func (rc *RankCommand) EventHandlers() []interface{} { return []interface{}{} }

func (rc *RankCommand) GetPermLevel() db.Permission {
	return db.PermAll
}
