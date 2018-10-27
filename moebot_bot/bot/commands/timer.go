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
		Mutex: sync.Mutex{},
		M:     make(map[string]*channelTimer),
	}
	return tc
}

func (tc *TimerCommand) Execute(pack *CommPackage) {
	channelID := pack.message.ChannelID
	if len(pack.params) > 0 {
		// Make sure the user has at least mod-level permissions before starting the timer
		if tc.Checker.HasPermission(pack.message.Author.ID, pack.member.Roles, pack.guild, db.PermMod) {
			tc.chTimers.Lock()

			if strings.EqualFold(pack.params[0], "start") {
				// If this channel timer is currently writing out, tell it to stop
				if chTimer, ok := tc.chTimers.M[channelID]; ok {
					// Stop existing writer
					if chTimer.requestCh != nil {
						close(chTimer.requestCh)
					}
				}

				// Create a new timer
				tc.chTimers.M[channelID] = &channelTimer{
					time:      time.Now(),
					requestCh: nil,
				}

				pack.session.ChannelMessageSend(pack.message.ChannelID, "Timer started!")
			} else if strings.EqualFold(pack.params[0], "stop") {
				if chTimer, ok := tc.chTimers.M[channelID]; ok {
					// Stop existing writer
					if chTimer.requestCh != nil {
						close(chTimer.requestCh)
					}
					delete(tc.chTimers.M, channelID)
					pack.session.ChannelMessageSend(pack.message.ChannelID, "Timer stopped.")
				} else {
					pack.session.ChannelMessageSend(pack.message.ChannelID, "No timer running.")
				}
			}

			tc.chTimers.Unlock()
		} else {
			pack.session.ChannelMessageSend(pack.message.ChannelID, pack.message.Author.Mention()+", you... you don't have permission to do that!")
		}
	} else {
		tc.chTimers.Lock()
		if chTimer, ok := tc.chTimers.M[channelID]; ok {
			// Close existing writer, then start a new one
			if chTimer.requestCh != nil {
				close(chTimer.requestCh)
			}
			chTimer.requestCh = make(chan string, 10)
			go writeTimes(pack, chTimer.time, chTimer.requestCh)
		} else {
			pack.session.ChannelMessageSend(pack.message.ChannelID, "No timer started for this channel...")
		}
		tc.chTimers.Unlock()
	}
}

func writeTimes(pack *CommPackage, startTime time.Time, reqCh <-chan string) {
	duration := time.Since(startTime)
	writes := 0
	currentWriteInterval := writeInterval

	// Write the time once right away
	go func() {
		pack.session.ChannelMessageSend(pack.message.ChannelID, fmtDuration(duration))
	}()
	writes++

	// Synchronize the writes to be divisible by the interval
	// (works well when interval is 5 so we get writes at times like 0:30, 0:35, 0:40, etc.)
	timeToSync := writeInterval - (duration % writeInterval)
	time.Sleep(timeToSync)
	duration += timeToSync

	// Set the interval to a short duration if we are going to do an "after-sync" write so that messages sent to the receiver will be handled first, and then do the write shortly afterwards.
	if timeToSync > time.Second {
		currentWriteInterval = 50 * time.Millisecond
	}

	// Start writing until we reach the max number of writes or get a message to stop
	for {
		select {
		case _, chOpen := <-reqCh:
			// Break out of this loop if the channel was closed
			if !chOpen {
				return
			}

		case <-time.After(currentWriteInterval):
			// Increment the duration and write time to the channel
			duration += currentWriteInterval
			go func() {
				pack.session.ChannelMessageSend(pack.message.ChannelID, fmtDuration(duration))
			}()
			writes++

			// Exit once we've reached the max write count
			if writes >= maxWrites {
				return
			}

			// Reset write interval in case it was changed to do an "after-sync" write
			currentWriteInterval = writeInterval
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
	return fmt.Sprintf("`%[1]s timer` - Checks the timer. Moderators may `start` or `stop` the timer", commPrefix)
}

type syncChannelTimerMap struct {
	sync.Mutex
	M map[string]*channelTimer
}

type channelTimer struct {
	time      time.Time
	requestCh chan string
}
