package commands

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/camd67/moebot/moebot_bot/bot/permissions"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

const (
	maxWrites     = 4               // Number of times to write out the time
	writeInterval = 5 * time.Second // Time between each write
)

type TimerCommand struct {
	chTimers syncChannelTimerMap
	Checker  permissions.PermissionChecker
}

func NewTimerCommand() *TimerCommand {
	tc := &TimerCommand{}
	tc.chTimers = syncChannelTimerMap{
		RWMutex: sync.RWMutex{},
		M:       make(map[string]*channelTimer),
	}
	return tc
}

func (tc *TimerCommand) Execute(pack *CommPackage) {
	channelID := pack.message.ChannelID
	if len(pack.params) > 0 && strings.EqualFold(pack.params[0], "start") {
		// Make sure the user has at least mod-level permissions before starting the timer
		if tc.Checker.HasPermission(pack.message.Author.ID, pack.member.Roles, pack.guild, db.PermMod) {
			tc.chTimers.Lock()

			// If this channel timer is currently writing out, tell it to stop
			if chTimer, ok := tc.chTimers.M[channelID]; ok {
				chTimer.Lock()
				if chTimer.isWriting {
					close(chTimer.requestCh)
				}
				chTimer.Unlock()
			}

			// Create a new timer
			tc.chTimers.M[channelID] = &channelTimer{
				time:      time.Now(),
				writes:    0,
				isWriting: false,
				requestCh: make(chan string, 10),
			}

			tc.chTimers.Unlock()
			pack.session.ChannelMessageSend(pack.message.ChannelID, "Timer started!")
		} else {
			pack.session.ChannelMessageSend(pack.message.ChannelID, pack.message.Author.Mention()+", you... you don't have permission to do that!")
		}
	} else {
		tc.chTimers.RLock()
		if chTimer, ok := tc.chTimers.M[channelID]; ok {
			chTimer.Lock()
			// Reset the number of writes
			chTimer.writes = 0

			// If the time is not writing, start it
			if !chTimer.isWriting {
				go chTimer.writeTimes(pack)
				chTimer.isWriting = true
			}
			chTimer.Unlock()
		} else {
			pack.session.ChannelMessageSend(pack.message.ChannelID, "No timer started for this channel...")
		}
		tc.chTimers.RUnlock()
	}
}

func (ct *channelTimer) writeTimes(pack *CommPackage) {
	duration := time.Since(ct.time)

	// Write the time once right away
	go func() {
		pack.session.ChannelMessageSend(pack.message.ChannelID, fmtDuration(duration))
		ct.Lock()
		ct.writes++
		ct.Unlock()
	}()

	// Synchronize the writes to be divisible by the interval (works well when interval is 5 so we get writes at times like 0:30, 0:35, 0:40, etc.)
	timeToSync := writeInterval - (duration % writeInterval)
	time.Sleep(timeToSync)
	duration += timeToSync

	// Write again if we spent a sufficient time syncing, otherwise just wait until the next write interval
	go func() {
		if timeToSync > time.Second {
			ct.Lock()
			pack.session.ChannelMessageSend(pack.message.ChannelID, fmtDuration(duration))
			ct.writes++
			ct.Unlock()
		}
	}()

	// Start writing until we reach the max number of writes or get a message to stop
	for {
		select {
		case _, chOpen := <-ct.requestCh:
			// Break out of this loop if the channel was closed
			if !chOpen {
				ct.Lock()
				ct.isWriting = false
				ct.Unlock()
				return
			}

		case <-time.After(writeInterval):
			// Increment the duration and write time to the channel
			duration += writeInterval
			go func() {
				pack.session.ChannelMessageSend(pack.message.ChannelID, fmtDuration(duration))
			}()

			// Exit once we've reached the max write count
			ct.Lock()
			ct.writes++
			if ct.writes >= maxWrites {
				ct.isWriting = false
				ct.Unlock()
				return
			}
			ct.Unlock()
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
	return fmt.Sprintf("`%[1]s timer` - Checks the timer. Moderators may provide the `start` option to start (or restart) the timer.", commPrefix)
}

type syncChannelTimerMap struct {
	sync.RWMutex
	M map[string]*channelTimer
}

type channelTimer struct {
	sync.Mutex
	time      time.Time
	writes    int
	isWriting bool
	requestCh chan string
}
