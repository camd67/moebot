package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type SchedulerFactory struct {
	session *discordgo.Session
}

func NewSchedulerFactory(session *discordgo.Session) *SchedulerFactory {
	return &SchedulerFactory{session: session}
}

func (f *SchedulerFactory) CreateScheduler(t db.SchedulerType) Scheduler {
	return NewChannelRotationScheduler(t, f.session)
}
