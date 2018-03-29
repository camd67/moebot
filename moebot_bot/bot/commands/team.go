package commands

import (
	"fmt"

	"github.com/camd67/moebot/moebot_bot/util/db"
)

type TeamCommand struct {
	Handler *RoleHandler
}

func (tc *TeamCommand) Execute(pack *CommPackage) {
	tc.Handler.processGuildRole([]string{"Nanachi", "Ozen", "Bondrewd", "Reg", "Riko", "Maruruk"}, pack.session, pack.params, pack.channel, pack.guild, pack.message, false, "team")
}

func (tc *TeamCommand) GetPermLevel() db.Permission {
	return db.PermAll
}

func (sc *TeamCommand) GetCommandKeys() []string {
	return []string{"TEAM"}
}

func (c *TeamCommand) GetCommandHelp(commPrefix string) string {
	return fmt.Sprintf("`%[1]s team <team name>` - Changes your team to one of the approved teams. `%[1]s team` to list all teams", commPrefix)
}
