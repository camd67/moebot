package db

import (
	"database/sql"
	"log"
)

type UserServerRank struct {
	Id       int
	ServerId int
	UserId   int
	Rank     int
}

type UserServerRankWrapper struct {
	UserUid   string
	ServerUid string
	Rank      int
}

const (
	userServerRankTable = `CREATE TABLE IF NOT EXISTS user_server_rank(
		Id SERIAL NOT NULL PRIMARY KEY,
		ServerId INTEGER NOT NULL REFERENCES server(id) ON DELETE CASCADE,
		UserId INTEGER NOT NULL REFERENCES user_profile(id) ON DELETE CASCADE,
		Rank INTEGER NOT NULL DEFAULT 0
	)`

	userServerRankQuery = `SELECT user_server_rank.Id, user_server_rank.ServerId, user_server_rank.UserId, user_server_rank.Rank FROM user_server_rank
		JOIN server ON server.Id = user_server_rank.Id
		JOIN user_profile ON user_profile.Id = user_server_rank.Id
		WHERE server.GuildUid = $1 AND user_profile.UserUid = $2`
	userServerRankQueryId = `SELECT Id, Rank FROM user_server_rank WHERE ServerId = $1 AND UserId = $2`
	userServerRankUpdate  = `UPDATE user_server_rank SET Rank = Rank + $2 WHERE Id = $1 RETURNING user_server_rank.Rank`
	userServerRankInsert  = `INSERT INTO user_server_rank(ServerId, UserId, Rank) VALUES ($1, $2, $3) RETURNING user_server_rank.Rank`
)

func UserServerRankQuery(userUid string, guildUid string) (usr *UserServerRank, err error) {
	row := moeDb.QueryRow(userServerRankQuery, guildUid, userUid)
	u := UserServerRank{}
	err = row.Scan(&u.Id, &u.ServerId, &u.UserId, &u.Rank)
	return &u, err
}

func UserServerRankUpdateOrInsert(userId int, serverId int, points int) (newPoint int, err error) {
	u := UserServerRank{
		ServerId: serverId,
		UserId:   userId,
	}
	row := moeDb.QueryRow(userServerRankQueryId, u.ServerId, u.UserId)
	if err = row.Scan(&u.Id, &u.Rank); err != nil {
		if err == sql.ErrNoRows {
			// no row, so insert it
			err = moeDb.QueryRow(userServerRankInsert, serverId, userId, points).Scan(&newPoint)
			if err != nil {
				log.Println("Error inserting userServerRank to db ", err)
				return
			}
			return
		}
	} else {
		// already have a row, update it
		err = moeDb.QueryRow(userServerRankUpdate, u.Id, points).Scan(&newPoint)
		if err != nil {
			log.Println("Error updating userServerRank", err)
			return
		}
	}
	return
}

func userServerRankCreateTable() {
	_, err := moeDb.Exec(userServerRankTable)
	if err != nil {
		log.Println("Error creating user server rank table", err)
		return
	}
}
