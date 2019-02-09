package db

import (
	"context"
	"database/sql"
	"log"

	"github.com/volatiletech/sqlboiler/boil"

	"github.com/volatiletech/sqlboiler/queries/qm"

	"github.com/camd67/moebot/moebot_bot/util/db/models"
)

func UserQueryOrInsert(userUid string) (u *models.UserProfile, err error) {
	u, e := models.UserProfiles(qm.Where(models.UserProfileColumns.UserUID+" = ?", userUid)).One(context.Background(), moeDb)
	if e != nil {
		if e == sql.ErrNoRows {
			u = &models.UserProfile{UserUID: userUid}
			e = u.Insert(context.Background(), moeDb, boil.Infer())
			if e != nil {
				log.Println("Error inserting user to db", e)
				return &models.UserProfile{}, e
			}
			return u, e
		}
	}
	// got a row, return it
	return
}
