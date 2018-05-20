package commands

import (
	"fmt"

	"github.com/camd67/moebot/moebot_bot/util/db"
	"github.com/camd67/moebot/moebot_bot/util/reddit"
)

type IrlCommand struct {
	RedditHandle *reddit.Handle
}

func (ic *IrlCommand) Execute(pack *CommPackage) {
	send, err := ic.RedditHandle.GetRandomImage("anime_irl")
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Ooops... Looks like this command isn't working right now. Sorry!")
		return
	}

	pack.session.ChannelMessageSendComplex(pack.channel.ID, send)
}

func (ic *IrlCommand) GetPermLevel() db.Permission {
	return db.PermAll
}
func (ic *IrlCommand) GetCommandKeys() []string {
	return []string{"IRL"}
}
func (ic *IrlCommand) GetCommandHelp(commPrefix string) string {
	return fmt.Sprintf("`%[1]s irl` - Posts a relatable meme.", commPrefix)
}
