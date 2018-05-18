package commands

import (
	"fmt"

	"github.com/camd67/moebot/moebot_bot/bot/permissions"
	"github.com/camd67/moebot/moebot_bot/util/db"
	"github.com/camd67/moebot/moebot_bot/util/reddit"
)

type MemeCommand struct {
	PermChecker  permissions.PermissionChecker
	RedditHandle *reddit.Handle
}

func (mc *MemeCommand) Execute(pack *CommPackage) {
	send, err := mc.RedditHandle.GetRandomImage("animemes")
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Ooops... Looks like this command isn't working right now. Sorry!")
		return
	}

	pack.session.ChannelMessageSendComplex(pack.channel.ID, send)
}

func (mc *MemeCommand) GetPermLevel() db.Permission {
	return db.PermAll
}
func (mc *MemeCommand) GetCommandKeys() []string {
	return []string{"MEME"}
}
func (mc *MemeCommand) GetCommandHelp(commPrefix string) string {
	return fmt.Sprintf("`%[1]s meme` - Posts a dank memerino.", commPrefix)
}
