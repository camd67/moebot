package commands

import (
	"strconv"
	"strings"

	"github.com/camd67/moebot/moebot_bot/util/db/types"
)

type EchoCommand struct {
}

func (ec *EchoCommand) Execute(pack *CommPackage) {
	_, err := strconv.Atoi(pack.params[0])
	if err != nil {
		pack.session.ChannelMessageSend(pack.message.ChannelID, "Sorry, that's an invalid channel ID")
		return
	}
	pack.session.ChannelMessageSend(pack.params[0], strings.Join(pack.params[1:], " "))
}

func (ec *EchoCommand) GetPermLevel() types.Permission {
	return types.PermMaster
}

func (ec *EchoCommand) GetCommandKeys() []string {
	return []string{"ECHO"}
}

func (ec *EchoCommand) GetCommandHelp(commPrefix string) string {
	return ""
}
