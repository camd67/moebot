package commands

import (
	"fmt"

	"github.com/camd67/moebot/moebot_bot/bot/permissions"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type HelpCommand struct {
	ComPrefix string
	Commands  func() []Command
	Checker   permissions.PermissionChecker
}

func (hc *HelpCommand) Execute(pack *CommPackage) {
	if len(pack.params) == 0 {
		message := "Moebot has the following commands:\n"
		for _, v := range hc.Commands() {
			if hc.Checker.HasPermission(pack.message.Author.ID, pack.member.Roles, pack.guild, v.GetPermLevel()) && v.GetCommandHelp(hc.ComPrefix) != "" {
				message += v.GetCommandHelp(hc.ComPrefix) + "\n"
			}
		}
		pack.session.ChannelMessageSend(pack.channel.ID, message)
	}
}

func (hc *HelpCommand) GetPermLevel() db.Permission {
	return db.PermAll
}

func (hc *HelpCommand) GetCommandKeys() []string {
	return []string{"HELP"}
}

func (hc *HelpCommand) GetCommandHelp(commPrefix string) string {
	return fmt.Sprintf("`%[1]s help` - Displays this message", commPrefix)
}
