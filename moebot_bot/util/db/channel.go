package db

import (
	"database/sql"
	"log"
)

type Channel struct {
	Id                int
	serverId          int
	ChannelUid        string
	BotAllowed        bool
	DefaultPinTarget  bool
	PinDestinationUid sql.NullString
}

const (
	channelTable = `CREATE TABLE IF NOT EXISTS channel(
		Id SERIAL NOT NULL PRIMARY KEY,
		serverId INTEGER REFERENCES server(Id) ON DELETE CASCADE,
		ChannelUid VARCHAR(20) NOT NULL,
		BotAllowed BOOLEAN NOT NULL DEFAULT TRUE
	)`

	channelQueryUid = `SELECT Id, serverId, ChannelUid, BotAllowed, PinDestinationUid, PinText FROM channel WHERE ChannelUid = $1`
	channelQueryId  = `SELECT Id, serverId, ChannelUid, BotAllowed, PinDestinationUid, PinText FROM channel WHERE Id = $1`

	channelInsert = `INSERT INTO channel (serverId, ChannelUid) VALUES($1, $2) RETURNING Id`
)

var channelUpdateTable = []string{
	"ALTER TABLE channel ADD COLUMN IF NOT EXISTS PinDestinationUid VARCHAR(20) NULL",
	"ALTER TABLE channel ADD COLUMN IF NOT EXISTS PinText BOOLEAN NOT NULL DEFAULT FALSE",
}

func ChannelQueryOrInsert(channelUid string, server *Server) (c *Channel, e error) {
	c = new(Channel)
	row := moeDb.QueryRow(channelQueryUid, channelUid)
	if e = row.Scan(&c.Id, &c.serverId, &c.ChannelUid, &c.BotAllowed, &c.DefaultPinTarget, &c.PinDestinationUid); e != nil {
		if e == sql.ErrNoRows {
			// no row, so insert it add in default values
			toInsert := &Channel{ChannelUid: channelUid, serverId: server.Id}
			e = moeDb.QueryRow(channelInsert, toInsert.serverId, toInsert.ChannelUid).Scan(&c.Id)
			if e != nil {
				log.Println("Error inserting channel to db ", e)
				return nil, e
			}
			return c, nil
		}
	}
	return c, nil
}

func ChannelQueryById(channelId int) (c *Channel, e error) {
	c = new(Channel)
	row := moeDb.QueryRow(channelQueryId, channelId)
	if e = row.Scan(&c.Id, &c.serverId, &c.ChannelUid, &c.BotAllowed, &c.DefaultPinTarget, &c.PinDestinationUid); e != nil {
		log.Println("Error querying channel", e)
		return nil, e
	}
	return c, nil
}

func channelCreateTable() {
	moeDb.Exec(channelTable)
	for _, alter := range channelUpdateTable {
		moeDb.Exec(alter)
	}
}
