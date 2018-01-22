package db

import (
	"database/sql"
	"log"
)

type User struct {
	Id      int
	UserUid string
}

const (
	userTable = `CREATE TABLE IF NOT EXISTS user(
		Id SERIAL NOT NULL PRIMARY KEY,
		UserUid VARCHAR(20) NOT NULL UNIQUE
	)`

	userQueryUid = `SELECT Id, UserUid FROM user WHERE UserUid = $1`
	userInsert   = `INSERT INTO user(UserUid) VALUES($1)`
)

func UserQueryOrInsert(userUid string) (u User, err error) {
	row := moeDb.QueryRow(userQueryUid, userUid)
	if err = row.Scan(&u.Id, &u.UserUid); err != nil {
		if err == sql.ErrNoRows {
			var insertId int
			err = moeDb.QueryRow(userInsert, userUid).Scan(&insertId)
			if err != nil {
				log.Println("Error inserting role to db", err)
				return
			}
			// for now we can just return back that ID since we already have the rest of the user
			return User{insertId, userUid}, nil
		}
	}
	// got a row, return it
	return
}

func userCreateTable() {
	_, err := moeDb.Exec(userTable)
	if err != nil {
		log.Println("Error creating user table", err)
		return
	}
}
