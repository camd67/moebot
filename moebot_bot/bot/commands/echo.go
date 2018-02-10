package commands

import (
	"strconv"
	"strings"

	"github.com/camd67/moebot/moebot_bot/util/db"

	"github.com/bwmarrin/discordgo"
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

func (ec *EchoCommand) Setup(session *discordgo.Session) {}

func (ec *EchoCommand) EventHandlers() []interface{} { return []interface{}{} }

func (ec *EchoCommand) GetPermLevel() db.Permission {
	return db.PermMaster
}
