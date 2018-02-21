package commands

import (
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type PollCommand struct {
	PollsHandler *PollsHandler
}

func (pc *PollCommand) Execute(pack *CommPackage) {
	if pack.params[0] == "-close" {
		pc.PollsHandler.closePoll(pack)
		return
	}
	var options []string
	var title string
	for i := 0; i < len(pack.params); i++ {
		if pack.params[i] == "-options" {
			options = parseOptions(pack.params[i+1:])
		}
		if pack.params[i] == "-title" {
			title = parseTitle(pack.params[i+1:])
		}
	}
	pc.PollsHandler.openPoll(pack, options, title)
}

func (pc *PollCommand) GetPermLevel() db.Permission {
	return db.PermMod
}

func (pc *PollCommand) GetCommandKeys() []string {
	return []string{"POLL"}
}
