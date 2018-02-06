package commands

import (
	"database/sql"
	"strconv"

	"github.com/camd67/moebot/moebot_bot/util"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type profile struct{}

func (p profile) execute(pack *CommPackage) {
	// technically we'll already have a user + server at this point, but may not have a usr. Still create if necessary
	_, err := db.ServerQueryOrInsert(pack.Guild.ID)
	_, err = db.UserQueryOrInsert(pack.Message.Author.ID)
	usr, err := db.UserServerRankQuery(pack.Message.Author.ID, pack.Guild.ID)
	if err != nil {
		if err != sql.ErrNoRows {
			pack.Session.ChannelMessageSend(pack.Message.ChannelID, "Sorry, there was an issue getting your information!")
			return
		}
		if err == sql.ErrNoRows || usr == nil {
			pack.Session.ChannelMessageSend(pack.Message.ChannelID, "Sorry, you don't have a rank yet!")
			return
		}
	}
	pack.Session.ChannelMessageSend(pack.Message.ChannelID, util.UserIdToMention(pack.Message.Author.ID)+"'s profile:\nRank: "+strconv.Itoa(usr.Rank))
}

func (p profile) setup() {}
