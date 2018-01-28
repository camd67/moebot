package db

import "log"

type UserServerRank struct {
	Id       int
	ServerId int
	UserId   int
	Rank     int
}

const (
	userServerRankTable = `CREATE TABLE IF NOT EXISTS user_server_rank(
		Id SERIAL NOT NULL PRIMARY KEY,
		ServerId INTEGER NOT NULL REFERENCES server(id) ON DELETE CASCADE,
		UserId INTEGER NOT NULL REFERENCES user_profile(id) ON DELETE CASCADE,
		Rank INTEGER NOT NULL DEFAULT 0
	)`
)

func userServerRankCreateTable() {
	_, err := moeDb.Exec(userServerRankTable)
	if err != nil {
		log.Println("Error creating user server rank table", err)
		return
	}
}
