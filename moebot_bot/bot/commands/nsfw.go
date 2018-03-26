package commands

import (
	"fmt"

	"github.com/camd67/moebot/moebot_bot/util/db"
)

type NsfwCommand struct {
	Handler *RoleHandler
}

func (nc *NsfwCommand) Execute(pack *CommPackage) {
	// force NSFW comm param so we can reuse guild role
	nc.Handler.processGuildRole([]string{"NSFW"}, pack.session, append([]string{"NSFW"}, pack.params...), pack.channel, pack.guild, pack.message, false, "nsfw")
}

func (nc *NsfwCommand) GetPermLevel() db.Permission {
	return db.PermAll
}

func (nc *NsfwCommand) GetCommandKeys() []string {
	return []string{"NSFW"}
}

func (c *NsfwCommand) GetCommandHelp(commPrefix string) string {
	return fmt.Sprintf("`%[1]s NSFW` - Gives you NSFW channel permissions", commPrefix)
}
