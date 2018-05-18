package commands

import (
	"fmt"

	"github.com/camd67/moebot/moebot_bot/bot/permissions"
	"github.com/camd67/moebot/moebot_bot/util/db"
	"github.com/camd67/moebot/moebot_bot/util/reddit"
)

type AwwnimeCommand struct {
	PermChecker  permissions.PermissionChecker
	RedditHandle *reddit.Handle
}

func (ac *AwwnimeCommand) Execute(pack *CommPackage) {
	send, err := ac.RedditHandle.GetRandomImage("awwnime")
	if err != nil {
		pack.session.ChannelMessageSend(pack.channel.ID, "Ooops... Looks like this command isn't working right now. Sorry!")
		return
	}

	pack.session.ChannelMessageSendComplex(pack.channel.ID, send)

	// request the image
	// send or something
	return
}

func (ac *AwwnimeCommand) GetPermLevel() db.Permission {
	return db.PermAll
}
func (ac *AwwnimeCommand) GetCommandKeys() []string {
	return []string{"RANDOM", "R"}
}
func (ac *AwwnimeCommand) GetCommandHelp(commPrefix string) string {
	return fmt.Sprintf("`%[1]s r` or `%[1]s random` - Posts a cute anime character.", commPrefix)
}
