package db

import (
	"database/sql"
	"log"
)

type Server struct {
	Id                  int
	GuildUid            string
	WelcomeMessage      sql.NullString
	RuleAgreement       sql.NullString
	DefaultPinChannelId sql.NullInt64
}

const (
	serverTable = `CREATE TABLE IF NOT EXISTS server(
		Id SERIAL NOT NULL PRIMARY KEY,
		GuildUid VARCHAR(20) NOT NULL UNIQUE,
		WelcomeMessage VARCHAR(500),
		RuleAgreement VARCHAR(50)
	)`

	serverQuery                = `SELECT Id, GuildUid, WelcomeMessage, RuleAgreement, DefaultPinChannelId FROM server WHERE Id = $1`
	serverQueryGuild           = `SELECT Id, GuildUid, WelcomeMessage, RuleAgreement, DefaultPinChannelId FROM server WHERE GuildUid = $1`
	serverInsert               = `INSERT INTO server(GuildUid, WelcomeMessage, RuleAgreement) VALUES ($1, $2, $3) RETURNING id`
	serverSetDefaultPinChannel = `UPDATE server SET DefaultPinChannelId = $1 WHERE Id = $2`
)

var serverUpdateTable = []string{
	"ALTER TABLE server ADD COLUMN IF NOT EXISTS DefaultPinChannelId INTEGER NULL",
}

func ServerQueryOrInsert(guildUid string) (s Server, e error) {
	row := moeDb.QueryRow(serverQueryGuild, guildUid)
	if e = row.Scan(&s.Id, &s.GuildUid, &s.WelcomeMessage, &s.RuleAgreement, &s.DefaultPinChannelId); e != nil {
		if e == sql.ErrNoRows {
			// no row, so insert it add in default values
			toInsert := Server{GuildUid: guildUid}
			var insertId int
			e = moeDb.QueryRow(serverInsert, toInsert.GuildUid, toInsert.WelcomeMessage, toInsert.RuleAgreement).Scan(&insertId)
			if e != nil {
				log.Println("Error inserting role to db ", e)
				return Server{}, e
			}
			row := moeDb.QueryRow(serverQuery, insertId)
			if e = row.Scan(&s.Id, &s.GuildUid, &s.WelcomeMessage, &s.RuleAgreement); e != nil {
				log.Println("Failed to read the newly inserted server row. This should pretty much never happen...", e)
				return Server{}, e
			}
			// normal flow of inserting a new row
			return s, e
		}
	}
	// normal flow of querying a row
	return
}

func ServerSetDefaultPinChannel(serverId int, channelId int) {
	moeDb.Exec(serverSetDefaultPinChannel, channelId, serverId)
}

func serverCreateTable() {
	moeDb.Exec(serverTable)
	for _, alter := range serverUpdateTable {
		moeDb.Exec(alter)
	}
}
