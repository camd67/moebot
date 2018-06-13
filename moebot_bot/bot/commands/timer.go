package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/camd67/moebot/moebot_bot/util/db"
)

type TimerCommand struct {
	StartTime time.Time
}

func (tc *TimerCommand) Execute(pack *CommPackage) {
	if len(pack.params) > 0 && strings.EqualFold(pack.params[0], "start") {
		tc.StartTime = time.Now()
		pack.session.ChannelMessageSend(pack.message.ChannelID, "Timer started OwO")
	} else {
		pack.session.ChannelMessageSend(pack.message.ChannelID, fmtDuration(time.Since(tc.StartTime)))
	}
}

func fmtDuration(dur time.Duration) string {
	remainingDur := dur.Round(time.Second)
	hours := remainingDur / time.Hour
	remainingDur -= hours * time.Hour
	minutes := remainingDur / time.Minute
	remainingDur -= minutes * time.Minute
	seconds := remainingDur / time.Second
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
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
