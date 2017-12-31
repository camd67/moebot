package db

type Channel struct {
	Id         int
	serverId   int
	ChannelUid string
	BotAllowed bool
}

const channelTable = `CREATE TABLE IF NOT EXISTS channel(
		Id SERIAL NOT NULL PRIMARY KEY,
		serverId INTEGER REFERENCES server(Id) ON DELETE CASCADE,
		ChannelUid VARCHAR(20) NOT NULL,
		BotAllowed BOOLEAN NOT NULL
	)`
