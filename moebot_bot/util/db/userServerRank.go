package db

import (
	"database/sql"
	"log"
	"strconv"
	"strings"

	"github.com/camd67/moebot/moebot_bot/util/db/types"
)

const (
	userServerRankTable = `CREATE TABLE IF NOT EXISTS user_server_rank(
		Id SERIAL NOT NULL PRIMARY KEY,
		ServerId INTEGER NOT NULL REFERENCES server(id) ON DELETE CASCADE,
		UserId INTEGER NOT NULL REFERENCES user_profile(id) ON DELETE CASCADE,
		Rank INTEGER NOT NULL DEFAULT 0,
		MessageSent BOOLEAN NOT NULL DEFAULT false
	)`

	userServerRankQuery = `SELECT user_server_rank.Id, user_server_rank.ServerId, user_server_rank.UserId, user_server_rank.Rank, user_server_rank.MessageSent FROM user_server_rank
		JOIN server ON server.Id = user_server_rank.ServerId
		JOIN user_profile ON user_profile.Id = user_server_rank.UserId
		WHERE server.GuildUid = $1 AND user_profile.UserUid = $2`
	userServerRankQueryId       = `SELECT Id, Rank, MessageSent FROM user_server_rank WHERE ServerId = $1 AND UserId = $2`
	userServerRankUpdate        = `UPDATE user_server_rank SET Rank = Rank + $2 WHERE Id = $1 RETURNING user_server_rank.Id, user_server_rank.Rank, user_server_rank.MessageSent`
	userServerRankInsert        = `INSERT INTO user_server_rank(ServerId, UserId, Rank) VALUES ($1, $2, $3) RETURNING user_server_rank.Id, user_server_rank.Rank, user_server_rank.MessageSent`
	userServerRankUpdateMessage = `UPDATE user_server_rank SET MessageSent = true WHERE Id = ANY ($1::integer[])`
)

func UserServerRankQuery(userUid string, guildUid string) (usr *types.UserServerRank, err error) {
	row := moeDb.QueryRow(userServerRankQuery, guildUid, userUid)
	u := types.UserServerRank{}
	err = row.Scan(&u.Id, &u.ServerId, &u.UserId, &u.Rank, &u.MessageSent)
	return &u, err
}

func UserServerRankUpdateOrInsert(userId int, serverId int, points int) (id int, newPoint int, messageSent bool, err error) {
	u := types.UserServerRank{
		ServerId: serverId,
		UserId:   userId,
	}
	row := moeDb.QueryRow(userServerRankQueryId, u.ServerId, u.UserId)
	if err = row.Scan(&u.Id, &u.Rank, &u.MessageSent); err != nil {
		if err == sql.ErrNoRows {
			// no row, so insert it
			err = moeDb.QueryRow(userServerRankInsert, serverId, userId, points).Scan(&id, &newPoint, &messageSent)
			if err != nil {
				log.Println("Error inserting userServerRank to db ", err)
				return
			}
			return
		}
	} else {
		// already have a row, update it
		err = moeDb.QueryRow(userServerRankUpdate, u.Id, points).Scan(&id, &newPoint, &messageSent)
		if err != nil {
			log.Println("Error updating userServerRank", err)
			return
		}
		return
	}
	return u.Id, u.Rank, u.MessageSent, nil
}

func UserServerRankSetMessageSent(entries []int) (err error) {
	ids := make([]string, len(entries))
	for i, e := range entries {
		ids[i] = strconv.Itoa(e)
	}
	idCollection := "{" + strings.Join(ids, ",") + "}"
	_, err = moeDb.Exec(userServerRankUpdateMessage, idCollection)
	if err != nil {
		log.Println("Error updating many user server ranks", err)
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
