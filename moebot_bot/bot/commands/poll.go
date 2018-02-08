package commands

import (
	"github.com/bwmarrin/discordgo"
)

type PollCommand struct {
	Checker      PermissionChecker
	PollsHandler PollsHandler
}

func (pc PollCommand) Execute(pack *CommPackage) {
	if !pc.Checker.HasModPerm(pack.message.Author.ID, pack.member.Roles) {
		pack.session.ChannelMessageSend(pack.channel.ID, "Sorry, this command has a minimum permission of mod")
		return
	}
	if pack.params[0] == "-close" {
		pc.PollsHandler.closePoll(pack)
		return
	}
	pc.PollsHandler.openPoll(pack)
}

func (pc PollCommand) Setup(session *discordgo.Session) {

}

func (pc PollCommand) EventHandlers() []interface{} {
	return []interface{}{pc.pollReactionsAdd}
}

func (pc PollCommand) pollReactionsAdd(session *discordgo.Session, reactionAdd *discordgo.MessageReactionAdd) {
	pc.PollsHandler.checkSingleVote(session, reactionAdd)
}
