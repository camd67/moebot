package commands

import (
	"fmt"
	"time"

	"github.com/camd67/moebot/moebot_bot/util/db/types"
)

type PingCommand struct {
}

func (pc *PingCommand) Execute(pack *CommPackage) {
	// seems this has some time drift when using docker for windows...
	messageTime, _ := pack.message.Timestamp.Parse()
	pingTime := time.Now().Sub(messageTime)
	pack.session.ChannelMessageSend(pack.channel.ID, "Latency to server: "+pingTime.String())
}

func (pc *PingCommand) GetPermLevel() types.Permission {
	return types.PermAll
}

func (pc *PingCommand) GetCommandKeys() []string {
	return []string{"PING"}
}

func (pc *PingCommand) GetCommandHelp(commPrefix string) string {
	return fmt.Sprintf("`%[1]s ping` - Checks the current latency to moebot's server and discord", commPrefix)
}
