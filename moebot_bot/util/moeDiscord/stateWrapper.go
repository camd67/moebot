/*
Extensions and additions to discordGo.

Anything that adds functionality to discordGo specifically should go in this package.
*/
package moeDiscord

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

func GetGuild(guildUid string, session *discordgo.Session) (*discordgo.Guild, error) {
	// session already caches guilds, but only guilds...
	// put here for consistency with other "gets"
	return session.Guild(guildUid)
}

func GetChannel(channelUid string, session *discordgo.Session) (channel *discordgo.Channel, err error) {
	channel, err = session.State.Channel(channelUid)
	// we only want to fetch the channel when we can't find it
	if err != nil && err != discordgo.ErrStateNotFound {
		channel, err = session.Channel(channelUid)
		if err != nil {
			log.Println("Error getting channel: "+channelUid, err)
			return nil, err
		}
		// fetched a valid channel, update the state and return it
		session.State.ChannelAdd(channel)
		return
	}
	// found a valid channel in the state
	return
}

func GetMember(memberUid string, guildUid string, session *discordgo.Session) (member *discordgo.Member, err error) {
	member, err = session.State.Member(guildUid, memberUid)
	// we only want to fetch the member when we can't find it
	if err != nil && err != discordgo.ErrStateNotFound {
		member, err = session.GuildMember(guildUid, memberUid)
		if err != nil {
			log.Println("Error getting member/guild: "+memberUid+"/"+guildUid, err)
			return nil, err
		}
		// fetched a valid member, update the state and return it
		session.State.MemberAdd(member)
		return
	}
	// found a valid member in the state
	return
}
