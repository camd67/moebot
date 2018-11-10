/*
Specific command functionality for moebot. Each command should be in its own file.
*/
package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util/db/types"
	"github.com/camd67/moebot/moebot_bot/util/event"
)

type CommPackage struct {
	session *discordgo.Session
	message *discordgo.Message
	guild   *discordgo.Guild
	member  *discordgo.Member
	channel *discordgo.Channel
	user    *types.UserProfile
	timer   *event.Timer
	params  []string
}

type Command interface {
	Execute(pack *CommPackage)
	GetPermLevel() types.Permission
	GetCommandKeys() []string
	GetCommandHelp(commPrefix string) string
}

type EventHandler interface {
	EventHandlers() []interface{}
}

type SetupHandler interface {
	Setup(session *discordgo.Session)
}

func NewCommPackage(session *discordgo.Session, message *discordgo.Message, guild *discordgo.Guild, member *discordgo.Member, channel *discordgo.Channel,
	params []string, user *types.UserProfile, timer *event.Timer) CommPackage {
	return CommPackage{
		session: session,
		message: message,
		guild:   guild,
		member:  member,
		channel: channel,
		user:    user,
		timer:   timer,
		params:  params,
	}
}
