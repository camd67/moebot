package commands

import (
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type RankCommand struct{}

func (rc *RankCommand) Execute(pack *CommPackage) {
	processGuildRole([]string{"Red", "Blue"}, pack.session, pack.params, pack.channel, pack.guild, pack.message, true)
}

func (rc *RankCommand) GetPermLevel() db.Permission {
	return db.PermAll
}

func (rc *RankCommand) GetCommandKeys() []string {
	return []string{"RANK"}
}
