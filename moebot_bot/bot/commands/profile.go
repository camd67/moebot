package commands

import (
	"database/sql"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/camd67/moebot/moebot_bot/util/db"
)

type ProfileCommand struct{}

func (pc *ProfileCommand) Execute(pack *CommPackage) {
	// technically we'll already have a user + server at this point, but may not have a usr. Still create if necessary
	server, err := db.ServerQueryOrInsert(pack.guild.ID)
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
	pack.session.ChannelMessageSend(pack.message.ChannelID, pack.message.Author.Mention()+"'s profile:\nRank: "+convertRankToString(usr.Rank,
		server.VeteranRank))
}

func convertRankToString(rank int, serverMax sql.NullInt64) (rankString string) {
	if !serverMax.Valid {
		// no server max? just give back the rank itself
		return strconv.Itoa(rank)
	}
	// naming strategy: every 1% till 10%, then every 2% until 30%, then every 3% until 60% then every 4% until 100%, then every 100% forever
	rankPrefixes := []string{"Newcomer", "Apprentice", "Rookie", "Regular", "Veteran"}
	percent := float64(rank) / float64(serverMax.Int64) * 100.0
	var rankPrefix string
	var rankSuffix int
	if percent < 10 {
		rankSuffix = int(percent / 2.0)
		rankPrefix = rankPrefixes[0]
	} else if percent < 30 {
		rankSuffix = int((percent - 10) / 4.0)
		rankPrefix = rankPrefixes[1]
	} else if percent < 60 {
		rankSuffix = int((percent - 30) / 6.0)
		rankPrefix = rankPrefixes[2]
	} else if percent < 100 {
		rankSuffix = int((percent - 60) / 8.0)
		rankPrefix = rankPrefixes[3]
	} else {
		rankSuffix = int((percent - 100) / 100)
		rankPrefix = rankPrefixes[4]
	}
	if rankSuffix != 0 {
		return rankPrefix + " " + strconv.Itoa(rankSuffix)
	}
	return rankPrefix
}

func (pc *ProfileCommand) Setup(session *discordgo.Session) {}

func (pc *ProfileCommand) EventHandlers() []interface{} { return []interface{}{} }

func (pc *ProfileCommand) GetPermLevel() db.Permission {
	return db.PermAll
}
