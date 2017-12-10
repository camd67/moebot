package db

type Channel struct {
	id         int
	serverId   int
	ChannelUid string
	BotAllowed bool
}

const channelTable = `CREATE TABLE IF NOT EXISTS channel(
		id SERIAL NOT NULL PRIMARY KEY,
		serverId INTEGER REFERENCES server(id) ON DELETE CASCADE,
		ChannelUid VARCHAR(20) NOT NULL,
		BotAllowed BOOLEAN NOT NULL
	)`
