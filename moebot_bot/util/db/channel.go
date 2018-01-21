package db

import (
	"database/sql"
	"log"
)

type Channel struct {
	Id           int
	serverId     int
	ChannelUid   string
	BotAllowed   bool
	MovePins     bool
	MoveTextPins bool
}

const (
	channelTable = `CREATE TABLE IF NOT EXISTS channel(
		Id SERIAL NOT NULL PRIMARY KEY,
		serverId INTEGER REFERENCES server(Id) ON DELETE CASCADE,
		ChannelUid VARCHAR(20) NOT NULL,
		BotAllowed BOOLEAN NOT NULL DEFAULT TRUE,
		MovePins BOOLEAN NOT NULL DEFAULT FALSE,
		MoveTextPins BOOLEAN NOT NULL DEFAULT FALSE
	)`

	channelQueryUid = `SELECT Id, serverId, ChannelUid, BotAllowed, MovePins, MoveTextPins FROM channel WHERE ChannelUid = $1`
	channelQueryId  = `SELECT Id, serverId, ChannelUid, BotAllowed, MovePins, MoveTextPins FROM channel WHERE Id = $1`

	channelInsert = `INSERT INTO channel (serverId, ChannelUid) VALUES($1, $2) RETURNING Id`

	channelSetPin = `UPDATE channel SET MovePins = $1, MoveTextPins = $2 WHERE Id = $3`
)

var channelUpdateTable = []string{
	"ALTER TABLE channel ADD COLUMN IF NOT EXISTS MovePins BOOLEAN NOT NULL DEFAULT FALSE",
	"ALTER TABLE channel ADD COLUMN IF NOT EXISTS MoveTextPins BOOLEAN NOT NULL DEFAULT FALSE",
}

func ChannelQueryOrInsert(channelUid string, server *Server) (c *Channel, e error) {
	c = new(Channel)
	row := moeDb.QueryRow(channelQueryUid, channelUid)
	if e = row.Scan(&c.Id, &c.serverId, &c.ChannelUid, &c.BotAllowed, &c.MovePins, &c.MoveTextPins); e != nil {
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
	if e = row.Scan(&c.Id, &c.serverId, &c.ChannelUid, &c.BotAllowed, &c.MovePins, &c.MoveTextPins); e != nil {
		log.Println("Error querying channel", e)
		return nil, e
	}
	return c, nil
}

func ChannelSetPin(channelId int, pinStatus bool, textPinStatus bool) error {
	//we want to always set textPinStatus to false if pinStatus is false
	_, err := moeDb.Exec(channelSetPin, pinStatus, textPinStatus && pinStatus, channelId)
	if err != nil {
		log.Println("Error setting pin channel status", err)
	}
	return err
}

func channelCreateTable() {
	moeDb.Exec(channelTable)
	for _, alter := range channelUpdateTable {
		moeDb.Exec(alter)
	}
}
