package db

import (
	"database/sql"
	"log"
)

type UserProfile struct {
	Id      int
	UserUid string
}

const (
	userProfileTable = `CREATE TABLE IF NOT EXISTS user_profile(
		Id SERIAL NOT NULL PRIMARY KEY,
		UserUid VARCHAR(20) NOT NULL UNIQUE
	)`

	userProfileQueryUid = `SELECT Id, UserUid FROM user_profile WHERE UserUid = $1`
	userProfileInsert   = `INSERT INTO user_profile(UserUid) VALUES($1)`
)

func UserQueryOrInsert(userUid string) (u UserProfile, err error) {
	row := moeDb.QueryRow(userProfileQueryUid, userUid)
	if err = row.Scan(&u.Id, &u.UserUid); err != nil {
		if err == sql.ErrNoRows {
			var insertId int
			err = moeDb.QueryRow(userProfileInsert, userUid).Scan(&insertId)
			if err != nil {
				log.Println("Error inserting role to db", err)
				return
			}
			// for now we can just return back that ID since we already have the rest of the user
			return UserProfile{insertId, userUid}, nil
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
