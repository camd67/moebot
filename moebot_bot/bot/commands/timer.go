package commands

import (
	"fmt"

	"github.com/camd67/moebot/moebot_bot/bot/permissions"

	"github.com/camd67/moebot/moebot_bot/util/db"
)

type TimerCommand struct {
	ComPrefix string
	Checker   permissions.PermissionChecker
}

func (tc *TimerCommand) GetPermLevel() db.Permission {
	return db.PermAll
}

func (tc *TimerCommand) GetCommandKeys() []string {
	return []string{"TIMER"}
}

func (tc *TimerCommand) GetCommandHelp(commPrefix string) string {
	return fmt.Sprintf("`%[1]s timer` - Checks the timestamp. Moderators may provide the `--start` option to begin start or restart the timer.", commPrefix)
}
