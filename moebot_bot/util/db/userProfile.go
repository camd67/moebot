package db

import (
	"database/sql"
	"log"

	"github.com/camd67/moebot/moebot_bot/util/db/types"
)

const (
	userProfileTable = `CREATE TABLE IF NOT EXISTS user_profile(
		Id SERIAL NOT NULL PRIMARY KEY,
		UserUid VARCHAR(20) NOT NULL UNIQUE
	)`

	userProfileQueryUid = `SELECT Id, UserUid FROM user_profile WHERE UserUid = $1`
	userProfileInsert   = `INSERT INTO user_profile(UserUid) VALUES($1) RETURNING Id`
)

func UserQueryOrInsert(userUid string) (u types.UserProfile, err error) {
	row := moeDb.QueryRow(userProfileQueryUid, userUid)
	if err = row.Scan(&u.Id, &u.UserUid); err != nil {
		if err == sql.ErrNoRows {
			var insertId int
			err = moeDb.QueryRow(userProfileInsert, userUid).Scan(&insertId)
			if err != nil {
				log.Println("Error inserting user to db", err)
				return
			}
			// for now we can just return back that ID since we already have the rest of the user
			return types.UserProfile{insertId, userUid}, nil
		}
	}
	// got a row, return it
	return
}

func userCreateTable() {
	_, err := moeDb.Exec(userProfileTable)
	if err != nil {
		log.Println("Error creating user table", err)
		return
	}
}
