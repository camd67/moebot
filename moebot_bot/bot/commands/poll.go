package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
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
	pc.PollsHandler.openPoll(pack)
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

func (pc *PollCommand) GetCommandKeys() []string {
	return []string{"POLL"}
}

func (pc *PollCommand) GetCommandHelp(commPrefix string) string {
	return fmt.Sprintf("`%[1]s poll -options <option 1, option 2, option 3, ...> -title <poll title>` - Master/All/Mod set up a poll with the given options. Type `%[1]s poll -close <poll id> to close`", commPrefix)
}
