package commands

import (
	"fmt"

	"github.com/camd67/moebot/moebot_bot/util/db"
)

type RankCommand struct {
	Handler *RoleHandler
}

func (rc *RankCommand) Execute(pack *CommPackage) {
	rc.Handler.processGuildRole([]string{"Red", "Blue"}, pack.session, pack.params, pack.channel, pack.guild, pack.message, true, "rank")
}

func (rc *RankCommand) GetPermLevel() db.Permission {
	return db.PermAll
}

func (rc *RankCommand) GetCommandKeys() []string {
	return []string{"RANK"}
}

func (c *RankCommand) GetCommandHelp(commPrefix string) string {
	return fmt.Sprintf("`%[1]s rank <rank name>` - Changes your rank to one of the approved ranks. `%[1]s rank` to list all the ranks", commPrefix)
}
