package db

import (
	"database/sql"
	"log"
)

type Channel struct {
	Id             int
	serverId       int
	ChannelUid     string
	BotAllowed     bool
	MovePins       bool
	MoveTextPins   bool
	DeletePin      bool
	MoveChannelUid sql.NullString
}

const (
	channelTable = `CREATE TABLE IF NOT EXISTS channel(
		Id SERIAL NOT NULL PRIMARY KEY,
		serverId INTEGER NOT NULL REFERENCES server(Id) ON DELETE CASCADE,
		ChannelUid VARCHAR(20) NOT NULL,
		BotAllowed BOOLEAN NOT NULL DEFAULT TRUE,
		MovePins BOOLEAN NOT NULL DEFAULT FALSE,
		MoveTextPins BOOLEAN NOT NULL DEFAULT FALSE,
		delete_pin BOOLEAN NOT NULL DEFAULT FALSE,
		move_channel_uid TEXT CHECK (char_length(move_channel_uid) < 21)
	)`

	channelQueryUid      = `SELECT Id, serverId, ChannelUid, BotAllowed, MovePins, MoveTextPins, delete_pin, move_channel_uid FROM channel WHERE ChannelUid = $1`
	channelQueryId       = `SELECT Id, serverId, ChannelUid, BotAllowed, MovePins, MoveTextPins, delete_pin, move_channel_uid FROM channel WHERE Id = $1`
	channelQueryServerId = `SELECT Id, serverId, ChannelUid, BotAllowed, MovePins, MoveTextPins, delete_pin, move_channel_uid FROM channel WHERE serverId = $1`

	channelInsert = `INSERT INTO channel (serverId, ChannelUid) VALUES($1, $2) RETURNING Id`

	channelUpdate = `UPDATE channel SET BotAllowed = $2, MovePins = $3, MoveTextPins = $4, delete_pin = $5, move_channel_uid = $6 WHERE Id = $1`
)

var channelUpdateTable = []string{
	"ALTER TABLE channel ADD COLUMN IF NOT EXISTS MovePins BOOLEAN NOT NULL DEFAULT FALSE",
	"ALTER TABLE channel ADD COLUMN IF NOT EXISTS MoveTextPins BOOLEAN NOT NULL DEFAULT FALSE",
	`ALTER TABLE channel ADD COLUMN IF NOT EXISTS move_channel_uid TEXT`,
	// Just as with server, this is dangerous to do as a live-update in production
	`ALTER TABLE channel DROP CONSTRAINT IF EXISTS channel_move_channel_uid_check`,
	`ALTER TABLE channel ADD CONSTRAINT channel_move_channel_uid_check CHECK (char_length(move_channel_uid) < 21)`,
	`ALTER TABLE channel ADD COLUMN IF NOT EXISTS delete_pin BOOLEAN NOT NULL DEFAULT FALSE`,
}

func ChannelQueryOrInsert(channelUid string, server *Server) (c *Channel, e error) {
	c = new(Channel)
	row := moeDb.QueryRow(channelQueryUid, channelUid)
	if e = row.Scan(&c.Id, &c.serverId, &c.ChannelUid, &c.BotAllowed, &c.MovePins, &c.MoveTextPins, &c.DeletePin, &c.MoveChannelUid); e != nil {
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

func ChannelUpdate(channel *Channel) (err error) {
	_, err = moeDb.Exec(channelUpdate, channel.Id, channel.BotAllowed, channel.MovePins, channel.MoveTextPins, channel.DeletePin, channel.MoveChannelUid)
	if err != nil {
		log.Println("Error update channel table", err)
		return
	}
	return
}

func ChannelQueryByServer(server Server) (channels []Channel, err error) {
	rows, err := moeDb.Query(channelQueryServerId, server.Id)
	if err != nil {
		log.Println("Error querying for channels", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var c Channel
		if err = rows.Scan(&c.Id, &c.serverId, &c.ChannelUid, &c.BotAllowed, &c.MovePins, &c.MoveTextPins, &c.DeletePin, &c.MoveChannelUid); err != nil {
			log.Println("Error scanning from channel table:", err)
			return
		}
		channels = append(channels, c)
	}
	return
}

func ChannelQueryById(channelId int) (c *Channel, e error) {
	c = new(Channel)
	row := moeDb.QueryRow(channelQueryId, channelId)
	if e = row.Scan(&c.Id, &c.serverId, &c.ChannelUid, &c.BotAllowed, &c.MovePins, &c.MoveTextPins, &c.DeletePin, &c.MoveChannelUid); e != nil {
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
