package commands

import (
	"database/sql"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type ProfileCommand struct{}

func (p ProfileCommand) Execute(pack *CommPackage) {
	// technically we'll already have a user + server at this point, but may not have a usr. Still create if necessary
	_, err := db.ServerQueryOrInsert(pack.guild.ID)
	_, err = db.UserQueryOrInsert(pack.message.Author.ID)
	usr, err := db.UserServerRankQuery(pack.message.Author.ID, pack.guild.ID)
	if err != nil {
		if err != sql.ErrNoRows {
			pack.session.ChannelMessageSend(pack.message.ChannelID, "Sorry, there was an issue getting your information!")
			return
		}
		if err == sql.ErrNoRows || usr == nil {
			pack.session.ChannelMessageSend(pack.message.ChannelID, "Sorry, you don't have a rank yet!")
			return
		}
	}
	pack.session.ChannelMessageSend(pack.message.ChannelID, util.UserIdToMention(pack.message.Author.ID)+"'s profile:\nRank: "+strconv.Itoa(usr.Rank))
}

func (p ProfileCommand) Setup(session *discordgo.Session) {}

func (p ProfileCommand) EventHandlers() []interface{} { return []interface{}{} }
