package db

import (
	"bytes"
	"database/sql"
	"log"
	"strconv"
	"sync"
)

/**
A Server (guild in discord terms) that stores information related to what over server settings are.
*/
type Server struct {
	Id             int
	GuildUid       string
	WelcomeMessage sql.NullString // Message user gets sent when they first join the server, either via PM or public message depending on WelcomeChannel
	RuleAgreement  sql.NullString // Message to type when the user agrees to the rules
	VeteranRank    sql.NullInt64  // Rank at which a user can be promoted to the veteran role
	VeteranRole    sql.NullString // Role to apply when a user can become a veteran
	BotChannel     sql.NullString // Where any bot related information or errors get sent to.
	Enabled        bool           // defaults to true, so that new servers that add moebot can immediately start using her. This can be turned off later
	WelcomeChannel sql.NullString // Channel to post a welcome message. If null, send via PM's
	StarterRole    sql.NullString // The role that is added when someone first joins a server
	BaseRole       sql.NullString // The role that is added when someone types the RuleAgreement message. Should only exist when RuleAgreement isn't null
}

const (
	serverTable = `CREATE TABLE IF NOT EXISTS server(
		Id SERIAL NOT NULL PRIMARY KEY,
		GuildUid VARCHAR(20) NOT NULL UNIQUE,
		WelcomeMessage VARCHAR(1900),
		RuleAgreement VARCHAR(1900),
		VeteranRank INTEGER,
		VeteranRole VARCHAR(20),
		BotChannel VARCHAR(20),
		Enabled BOOLEAN NOT NULL DEFAULT TRUE,
		WelcomeChannel VARCHAR(20),
		StarterRole VARCHAR(20),
		BaseRole VARCHAR(20)
	)`

	serverColumnNames        = `GuildUid, WelcomeMessage, RuleAgreement, VeteranRank, VeteranRole, BotChannel, Enabled, WelcomeChannel, StarterRole, BaseRole`
	serverInsertColumnNames  = `GuildUid, WelcomeMessage, RuleAgreement, VeteranRank, VeteranRole, BotChannel, WelcomeChannel, StarterRole, BaseRole`
	serverInsertColumnParams = `$1, $2, $3, $4, $5, $6, $7, $8, $9`
	serverSetParams          = `WelcomeMessage = $2, RuleAgreement = $3, VeteranRank = $4, VeteranRole = $5, BotChannel = $6, Enabled = $7, StarterRole = $8, BaseRole = $9, WelcomeChannel = $10`

	serverQuery      = `SELECT Id, ` + serverColumnNames + ` FROM server WHERE Id = $1`
	serverQueryGuild = `SELECT Id, ` + serverColumnNames + ` FROM server WHERE GuildUid = $1`
	serverInsert     = `INSERT INTO server(` + serverInsertColumnNames + `) VALUES (` + serverInsertColumnParams + `) RETURNING id`
	serverUpdate     = `UPDATE server SET ` + serverSetParams + ` WHERE Id = $1`
)

var (
	serverUpdateTable = []string{
		`ALTER TABLE server ADD COLUMN IF NOT EXISTS VeteranRank INTEGER`,
		`ALTER TABLE server ADD COLUMN IF NOT EXISTS VeteranRole VARCHAR(20)`,
		`ALTER TABLE server ADD COLUMN IF NOT EXISTS BotChannel VARCHAR(20)`,
		`ALTER TABLE server ADD COLUMN IF NOT EXISTS Enabled BOOLEAN NOT NULL DEFAULT TRUE`,
		`ALTER TABLE server ADD COLUMN IF NOT EXISTS WelcomeChannel VARCHAR(20)`,
		`ALTER TABLE server ADD COLUMN IF NOT EXISTS StarterRole VARCHAR(20)`,
		`ALTER TABLE server ADD COLUMN IF NOT EXISTS BaseRole VARCHAR(20)`,
		`ALTER TABLE server DROP CONSTRAINT IF EXISTS server_defaultpinchannelid_fkey`,
		`ALTER TABLE server DROP COLUMN IF EXISTS DefaultPinChannelId`,
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
	if e = serverScan(row, &s); e != nil {
		if e == sql.ErrNoRows {
			// no row, so insert it add in default values
			toInsert := Server{GuildUid: guildUid}
			var insertId int
			e = moeDb.QueryRow(serverInsert, &toInsert.GuildUid, &s.WelcomeMessage, &s.RuleAgreement, &s.VeteranRank, &s.VeteranRole,
				&s.BotChannel, &s.WelcomeChannel, &s.StarterRole, &s.BaseRole).Scan(&insertId)
			if e != nil {
				log.Println("Error inserting role to db ", e)
				return Server{}, e
			}
			row := moeDb.QueryRow(serverQuery, insertId)
			if e = serverScan(row, &s); e != nil {
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

func serverScan(row *sql.Row, s *Server) error {
	return row.Scan(&s.Id, &s.GuildUid, &s.WelcomeMessage, &s.RuleAgreement, &s.VeteranRank, &s.VeteranRole, &s.BotChannel, &s.Enabled,
		&s.WelcomeChannel, &s.StarterRole, &s.BaseRole)
}

func ServerSprint(s Server) (out string) {
	var buf bytes.Buffer
	buf.WriteString("Server: ")
	if s.WelcomeMessage.Valid {
		buf.WriteString("{WelcomeMessage: `")
		if len(s.WelcomeMessage.String) > 25 {
			buf.WriteString(s.WelcomeMessage.String[0:25])
			buf.WriteString("...")
		} else {
			buf.WriteString(s.WelcomeMessage.String)
		}
		buf.WriteString("`}")
	}
	if s.WelcomeChannel.Valid {
		if !s.WelcomeMessage.Valid {
			buf.WriteString("{!!! MISCONFIG !!!: `welcome channel but no message!`}")
		}
		buf.WriteString("{WelcomeChannel: `")
		buf.WriteString(s.WelcomeChannel.String)
		buf.WriteString("`}")
	} else {
		if s.WelcomeMessage.Valid {
			buf.WriteString("{WelcomeChannel: `Sent via DM`}")
		}
	}
	if s.RuleAgreement.Valid {
		buf.WriteString("{RuleAgreement: `")
		if len(s.RuleAgreement.String) > 25 {
			buf.WriteString(s.RuleAgreement.String[0:25])
			buf.WriteString("...")
		} else {
			buf.WriteString(s.RuleAgreement.String)
		}
		buf.WriteString("`}")
		if !s.BaseRole.Valid {
			buf.WriteString("{!!! MISCONFIG !!!: `Rule agreement found but no base role set`}")
		}
	}
	if s.BotChannel.Valid {
		buf.WriteString("{BotChannel: `")
		buf.WriteString(s.BotChannel.String)
		buf.WriteString("`}")
	}
	if s.StarterRole.Valid {
		buf.WriteString("{StarterRole: `")
		buf.WriteString(s.StarterRole.String)
		buf.WriteString("`}")
	}
	if s.BaseRole.Valid {
		buf.WriteString("{BaseRole: `")
		buf.WriteString(s.BaseRole.String)
		buf.WriteString("`}")
	}
	if s.Enabled {
		buf.WriteString("{Enabled: `")
		buf.WriteString(strconv.FormatBool(s.Enabled))
		buf.WriteString("`}")
	}
	if s.VeteranRank.Valid {
		buf.WriteString("{VeteranRank: `")
		buf.WriteString(strconv.Itoa(int(s.VeteranRank.Int64)))
		buf.WriteString("`}")
		if !s.VeteranRole.Valid {
			buf.WriteString("{!!! MISCONFIG !!!: `veteran rank provided but no role provided!`}")
		}
	}
	if s.VeteranRole.Valid {
		buf.WriteString("{VeteranRole: `")
		buf.WriteString(s.VeteranRole.String)
		buf.WriteString("`}")
		if !s.VeteranRank.Valid {
			buf.WriteString("{!!! MISCONFIG !!!: `veteran role provided but no rank provided!`}")
		}
	}
	return buf.String()
}

func FlushServerCache() {
	serverMemoryBuffer.Lock()
	defer serverMemoryBuffer.Unlock()
	serverMemoryBuffer.m = make(map[string]Server)
}

func ServerFullUpdate(s Server) (err error) {
	_, err = moeDb.Exec(serverUpdate, s.Id, s.WelcomeMessage, s.RuleAgreement, s.VeteranRank, s.VeteranRole, s.BotChannel, s.Enabled,
		s.StarterRole, s.BaseRole, s.WelcomeChannel)
	if err != nil {
		log.Println("There was an error updating the server table", err)
	}
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
