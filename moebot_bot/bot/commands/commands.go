package commands

import (
	"github.com/bwmarrin/discordgo"
)

type CommPackage struct {
	session *discordgo.Session
	message *discordgo.Message
	guild   *discordgo.Guild
	member  *discordgo.Member
	channel *discordgo.Channel
	params  []string
}

type Command interface {
	Execute(pack *CommPackage)
	Setup(session *discordgo.Session)
	EventHandlers() []interface{}
}

func NewCommPackage(session *discordgo.Session, message *discordgo.Message, guild *discordgo.Guild, member *discordgo.Member, channel *discordgo.Channel, params []string) CommPackage {
	return CommPackage{
		session: session,
		message: message,
		guild:   guild,
		member:  member,
		channel: channel,
		params:  params,
	}
}
