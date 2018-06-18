package commands

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/camd67/moebot/moebot_bot/bot/permissions"
	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

const (
	maxWrites = 5 // Number of times to write out the time
)

type TimerCommand struct {
	chTimers util.SyncChannelTimerMap
	Checker  permissions.PermissionChecker
}

func NewTimerCommand() *TimerCommand {
	tc := &TimerCommand{}
	tc.chTimers = util.SyncChannelTimerMap{
		RWMutex: sync.RWMutex{},
		M:       make(map[string]time.Time),
	}
	return tc
}

func (tc *TimerCommand) Execute(pack *CommPackage) {
	channelID := pack.message.ChannelID
	if len(pack.params) > 0 && strings.EqualFold(pack.params[0], "start") {
		if tc.Checker.HasPermission(pack.message.Author.ID, pack.member.Roles, pack.guild, db.PermMod) {
			tc.chTimers.Lock()
			tc.chTimers.M[channelID] = time.Now()
			tc.chTimers.Unlock()
			pack.session.ChannelMessageSend(pack.message.ChannelID, "Timer started!")
		} else {
			pack.session.ChannelMessageSend(pack.message.ChannelID, pack.message.Author.Mention()+", you... you don't have permission to do that!")
		}
	} else {
		tc.chTimers.RLock()
		if v, ok := tc.chTimers.M[channelID]; ok {
			go tc.writeTimesToChannel(pack, time.Since(v))
		} else {
			pack.session.ChannelMessageSend(pack.message.ChannelID, "No timer started for this channel...")
		}
		tc.chTimers.RUnlock()
	}
}

func (tc *TimerCommand) writeTimesToChannel(pack *CommPackage, startDuration time.Duration) {
	//Write the time once right away
	pack.session.ChannelMessageSend(pack.message.ChannelID, fmtDuration(startDuration))
	duration := startDuration
	writes := 1
	for {
		select {
		case <-time.After(time.Second * 1):
			duration += time.Duration(time.Second)
			go func() {
				pack.session.ChannelMessageSend(pack.message.ChannelID, fmtDuration(duration))
			}()

			// Exit once we've reached the max write count
			writes++
			if writes >= maxWrites {
				return
			}
		}
	}
}

// fmtDuration formats a duration into a hh:mm:ss format
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
