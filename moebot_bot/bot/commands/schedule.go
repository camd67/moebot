package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"

	"github.com/camd67/moebot/moebot_bot/util/db"
)

type ScheduleCommand struct {
	schedulers []Scheduler
}

func NewScheduleCommand(session *discordgo.Session) *ScheduleCommand {
	return &ScheduleCommand{
		schedulers: []Scheduler{
			NewChannelRotationScheduler(db.SchedulerChannelRotation, session),
		},
	}
}

func (c *ScheduleCommand) Execute(pack *CommPackage) {

}
func (c *ScheduleCommand) GetPermLevel() db.Permission {
	return db.PermMod
}
func (c *ScheduleCommand) GetCommandKeys() []string {
	return []string{"SCHEDULE"}
}
func (c *ScheduleCommand) GetCommandHelp(commPrefix string) string {
	return fmt.Sprintf("`%[1]s schedule <schedule type> <options>` - Master/Mod Creates a new scheduler with the given type and options. `%[1]s schedule` to list available schedulers.", commPrefix)
}
