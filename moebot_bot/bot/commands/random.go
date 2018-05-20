package commands

import (
	"fmt"

	"github.com/camd67/moebot/moebot_bot/util/db"
	"github.com/camd67/moebot/moebot_bot/util/reddit"
)

type RandomCommand struct {
	RedditHandle *reddit.Handle
}

func (ac *RandomCommand) Execute(pack *CommPackage) {
	send, err := ac.RedditHandle.GetRandomImage("awwnime")
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Ooops... Looks like this command isn't working right now. Sorry!")
		return
	}

	pack.session.ChannelMessageSendComplex(pack.channel.ID, send)
}

func (ac *RandomCommand) GetPermLevel() db.Permission {
	return db.PermAll
}
func (ac *RandomCommand) GetCommandKeys() []string {
	return []string{"RANDOM", "R"}
}
func (ac *RandomCommand) GetCommandHelp(commPrefix string) string {
	return fmt.Sprintf("`%[1]s random` - Posts a cute anime character.", commPrefix)
}
