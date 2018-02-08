package commands

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

type PingCommand struct {
}

func (pc PingCommand) Execute(pack *CommPackage) {
	// seems this has some time drift when using docker for windows... need to verify if it's accurate on the server
	messageTime, _ := pack.message.Timestamp.Parse()
	pingTime := time.Duration(time.Now().UnixNano() - messageTime.UnixNano())
	pack.session.ChannelMessageSend(pack.channel.ID, "Latency to server: "+pingTime.String())
}

func (pc PingCommand) Setup(session *discordgo.Session) {}

func (pc PingCommand) EventHandlers() []interface{} { return []interface{}{} }
