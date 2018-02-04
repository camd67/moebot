package db

import (
	"database/sql"
	"log"
	"sync"
)

type Server struct {
	Id                  int
	GuildUid            string
	WelcomeMessage      sql.NullString
	RuleAgreement       sql.NullString
	VeteranRank         sql.NullInt64
	VeteranRole         sql.NullString
	BotChannel          sql.NullString
	DefaultPinChannelId sql.NullInt64
}

const (
	serverTable = `CREATE TABLE IF NOT EXISTS server(
		Id SERIAL NOT NULL PRIMARY KEY,
		GuildUid VARCHAR(20) NOT NULL UNIQUE,
		WelcomeMessage VARCHAR(500),
		RuleAgreement VARCHAR(50),
		VeteranRank INTEGER,
		VeteranRole VARCHAR(20),
		DefaultPinChannelId INTEGER NULL REFERENCES channel(Id),
		BotChannel VARCHAR(20)
	)`

	serverQuery      = `SELECT Id, GuildUid, WelcomeMessage, RuleAgreement, VeteranRank, VeteranRole, DefaultPinChannelId, BotChannel FROM server WHERE Id = $1`
	serverQueryGuild = `SELECT Id, GuildUid, WelcomeMessage, RuleAgreement, VeteranRank, VeteranRole, DefaultPinChannelId, BotChannel FROM server WHERE GuildUid = $1`
	serverInsert     = `INSERT INTO server(GuildUid, WelcomeMessage, RuleAgreement, VeteranRank, VeteranRole, DefaultPinChannelId, BotChannel) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	serverUpdate     = `UPDATE server SET WelcomeMessage = $2, RuleAgreement = $3, VeteranRank = $4, VeteranRole = $5, DefaultPinChannelId = $6, BotChannel = $7 WHERE Id = $1`
)

var (
	serverUpdateTable = []string{
		`ALTER TABLE server ADD COLUMN IF NOT EXISTS VeteranRank INTEGER`,
		`ALTER TABLE server ADD COLUMN IF NOT EXISTS VeteranRole VARCHAR(20)`,
		"ALTER TABLE server ADD COLUMN IF NOT EXISTS DefaultPinChannelId INTEGER NULL REFERENCES channel(Id)",
		`ALTER TABLE server ADD COLUMN IF NOT EXISTS BotChannel VARCHAR(20)`,
	}

	serverMemoryBuffer = struct {
		sync.RWMutex
		m map[string]Server
	}{m: make(map[string]Server)}
)

func ServerQueryOrInsert(guildUid string) (s Server, e error) {
	serverMemoryBuffer.RLock()
	if memServer, ok := serverMemoryBuffer.m[guildUid]; ok {
		serverMemoryBuffer.RUnlock()
		return memServer, nil
	}
	serverMemoryBuffer.RUnlock()
	row := moeDb.QueryRow(serverQueryGuild, guildUid)
	if e = row.Scan(&s.Id, &s.GuildUid, &s.WelcomeMessage, &s.RuleAgreement, &s.VeteranRank, &s.VeteranRole, &s.DefaultPinChannelId, &s.BotChannel); e != nil {
		if e == sql.ErrNoRows {
			// no row, so insert it add in default values
			toInsert := Server{GuildUid: guildUid}
			var insertId int
			e = moeDb.QueryRow(serverInsert, toInsert.GuildUid, toInsert.WelcomeMessage, toInsert.RuleAgreement, &s.VeteranRank, &s.VeteranRole, &s.BotChannel).Scan(&insertId)
			if e != nil {
				log.Println("Error inserting role to db ", e)
				return Server{}, e
			}
			row := moeDb.QueryRow(serverQuery, insertId)
			if e = row.Scan(&s.Id, &s.GuildUid, &s.WelcomeMessage, &s.RuleAgreement, &s.VeteranRank, &s.VeteranRole, &s.BotChannel); e != nil {
				log.Println("Failed to read the newly inserted server row. This should pretty much never happen...", e)
				return Server{}, e
			}
			// normal flow of inserting a new row
			serverMemoryBuffer.Lock()
			serverMemoryBuffer.m[guildUid] = s
			serverMemoryBuffer.Unlock()
			return s, e
		}
	}
	// normal flow of querying a row
	return
}

func FlushServerCache() {
	serverMemoryBuffer.Lock()
	defer serverMemoryBuffer.Unlock()
	serverMemoryBuffer.m = make(map[string]Server)
}

func ServerFullUpdate(s Server) (err error) {
	_, err = moeDb.Exec(serverUpdate, s.Id, s.WelcomeMessage, s.RuleAgreement, s.VeteranRank, s.VeteranRole, s.BotChannel)
	return
}

func serverCreateTable() {
	_, err := moeDb.Exec(serverTable)
	if err != nil {
		log.Println("Error creating server table", err)
		return
	}
	for _, alter := range serverUpdateTable {
		_, err = moeDb.Exec(alter)
		if err != nil {
			log.Println("Error alterting server table", err)
			return
		}
	}
}

func ServerSetDefaultPinChannel(serverId int, channelId int) error {
	_, err := moeDb.Exec(serverSetDefaultPinChannel, channelId, serverId)
	if err != nil {
		log.Println("Failed to set pin channel", err)
	}
	return err
}
