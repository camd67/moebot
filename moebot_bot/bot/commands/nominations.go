package commands

import "github.com/camd67/moebot/moebot_bot/util/db"

type NominationsCommand struct {
	PollsHandler *PollsHandler
}

func (c *NominationsCommand) Execute(commPack *CommPackage) {
	if len(commPack.params) != 2 {
		commPack.session.ChannelMessageSend(commPack.channel.ID, "Sorry, you need to specify either `open` or `close` and the name for the nomination.")
		return
	}
	switch commPack.params[0] {
	case "open":
		c.openNominations(commPack, commPack.params[1])
	case "close":
		c.closeNominations(commPack, commPack.params[1])
	default:
		commPack.session.ChannelMessageSend(commPack.channel.ID, "Sorry, you need to specify either `open` or `close` and the name for the nomination.")
		break
	}
}

func (c *NominationsCommand) openNominations(commPack *CommPackage, nominationsName string) {
	
}

func (c *NominationsCommand) closeNominations(commPack *CommPackage, nominationsName string) {
	var nominationsList []string
	c.PollsHandler.openPoll(commPack, nominationsList, nominationsName)
}

func (c *NominationsCommand) GetPermLevel() db.Permission {
	return db.PermMod
}

func (c *NominationsCommand) GetCommandKeys() []string {
	return []string{"NOMINATIONS"}
}
