package commands

import (
	"fmt"
	"time"

	"github.com/camd67/moebot/moebot_bot/util/db"
)

type TimerCommand struct {
	StartTime time.Time
}

func (tc *TimerCommand) Execute(pack *CommPackage) {
	pack.session.ChannelMessageSend(pack.message.ChannelID, time.Since(tc.StartTime).Round(time.Second).String())
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
