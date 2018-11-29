package db

import (
	"context"
	"database/sql"
	"log"

	"github.com/volatiletech/sqlboiler/boil"

	"github.com/camd67/moebot/moebot_bot/util/db/models"
	"github.com/volatiletech/sqlboiler/queries/qm"
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

func UserServerRankQuery(userUid string, guildUid string) (usr *models.UserServerRank, err error) {
	return models.UserServerRanks(
		qm.InnerJoin("server s on s.id = user_server_rank.server_id"),
		qm.InnerJoin("user_profile u on u.id = user_server_rank.user_id"),
		qm.Where("guild_uid = ? AND user_uid = ?", guildUid, userUid),
	).One(context.Background(), moeDb)
}

func UserServerRankUpdateOrInsert(userId int, serverId int, points int) (id int, newPoint int, messageSent bool, err error) {
	u, err := models.UserServerRanks(qm.Where("server_id = ? AND user_id = ?", userId, userId)).One(context.Background(), moeDb)
	if err == sql.ErrNoRows {
		u = &models.UserServerRank{
			ServerID: serverId,
			UserID:   userId,
			Rank:     points,
		}
		err = u.Insert(context.Background(), moeDb, boil.Infer())
		if err != nil {
			log.Println("Error inserting userServerRank to db ", err)
			return
		}
	} else {
		u.Rank += points
		u.Update(context.Background(), moeDb, boil.Infer())
		if err != nil {
			log.Println("Error updating userServerRank", err)
			return
		}
	}
	return u.ID, u.Rank, u.MessageSent, nil
}

func UserServerRankSetMessageSent(entries []int) (err error) {
	convertedIds := make([]interface{}, len(entries))
	for index, num := range entries {
		convertedIds[index] = num
	}
	_, err = models.UserServerRanks(qm.WhereIn("id in ?", convertedIds...)).UpdateAll(context.Background(), moeDb, models.M{"message_sent": true})
	if err != nil {
		log.Println("Error updating many user server ranks", err)
	}
	return
}
