package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type PollCommand struct {
	Checker      *PermissionChecker
	PollsHandler *PollsHandler
}

func (pc *PollCommand) Execute(pack *CommPackage) {
	if pack.params[0] == "-close" {
		pc.PollsHandler.closePoll(pack)
		return
	}
	pc.PollsHandler.openPoll(pack)
}

func (pc *PollCommand) Setup(session *discordgo.Session) {

}

func (pc *PollCommand) EventHandlers() []interface{} {
	return []interface{}{pc.pollReactionsAdd}
}

func (pc *PollCommand) pollReactionsAdd(session *discordgo.Session, reactionAdd *discordgo.MessageReactionAdd) {
	pc.PollsHandler.checkSingleVote(session, reactionAdd)
}

func (pc *PollCommand) GetPermLevel() db.Permission {
	return db.PermMod
}
