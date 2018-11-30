package db

import (
	"context"
	"database/sql"
	"log"

	"github.com/volatiletech/sqlboiler/boil"

	"github.com/camd67/moebot/moebot_bot/util/db/models"
	"github.com/volatiletech/sqlboiler/queries/qm"
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
