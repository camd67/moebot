package db

import (
	"context"
	"database/sql"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"

	"github.com/camd67/moebot/moebot_bot/util/db/models"
)

var (
	serverMemoryBuffer = struct {
		sync.RWMutex
		m map[string]models.Server
	}{m: make(map[string]models.Server)}
)

func ServerQueryOrInsert(guildUid string) (s *models.Server, e error) {
	if s, e := models.Servers(qm.Where(models.ServerColumns.GuildUID+" = ?", guildUid)).One(context.Background(), moeDb); e != nil {
		if e == sql.ErrNoRows {
			// no row, so insert it add in default values
			s = &models.Server{GuildUID: guildUid, Enabled: true}
			e = s.Insert(context.Background(), moeDb, boil.Infer())
			if e != nil {
				log.Println("Error inserting role to db ", e)
				return &models.Server{}, e
			}
			return s, e
		}
	}
	// normal flow of querying a row
	return
}

func ServerQueryById(id int) (s *models.Server, e error) {
	if s, e = models.FindServer(context.Background(), moeDb, id); e != nil {
		log.Println("Error retrieving the server from db by id", e)
		return &models.Server{}, e
	}
	return s, e
}

func ServerQueryByGuildUid(guildUid string) (s *models.Server, e error) {
	if s, e = models.Servers(qm.Where("guild_uid = ?", guildUid)).One(context.Background(), moeDb); e != nil {
		log.Println("Error retrieving the server from db by guild uid", e)
		return &models.Server{}, e
	}
	return s, e
}

func ServerSprint(s *models.Server) (out string) {
	var buf strings.Builder
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
		buf.WriteString(strconv.Itoa(s.VeteranRank.Int))
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
	serverMemoryBuffer.m = make(map[string]models.Server)
}

func ServerFullUpdate(s *models.Server) (err error) {
	_, err = s.Update(context.Background(), moeDb, boil.Infer())
	if err != nil {
		log.Println("There was an error updating the server table", err)
	}
	return
}
