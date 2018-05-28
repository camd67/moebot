package db

import (
	"database/sql"
	"log"
)

const (
	guildEventCreateTable = `CREATE TABLE IF NOT EXISTS guildEvent(
		ID SERIAL NOT NULL PRIMARY KEY,
		Description VARCHAR NOT NULL,
		Date TIME NOT NULL,
		GuildUID VARCHAR(20) NOT NULL
	)`
	guildEventSelect = `SELECT Password FROM users WHERE Username = $1`
	guildEventInsert = `INSERT INTO users (Username, Password, Email, DiscordUID, DiscordAuthToken, DiscordRefreshToken, DiscordTokenExpiration) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING ID`
)

type guildEventTable struct {
	db *sql.DB
}

func (t *guildEventTable) createTable() {
	_, err := t.db.Exec(guildEventCreateTable)
	if err != nil {
		log.Fatalln(err)
	}
}
