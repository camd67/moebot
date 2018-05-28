package db

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"

	"github.com/lib/pq"

	"golang.org/x/crypto/bcrypt"
)

const (
	userCreateTable = `CREATE TABLE IF NOT EXISTS users(
		ID UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		Username VARCHAR NOT NULL UNIQUE,
		Password VARCHAR NULL,
		Email VARCHAR NULL,
		DiscordUID VARCHAR(20) NULL,
		DiscordAuthToken VARCHAR NULL,
		DiscordRefreshToken VARCHAR NULL,
		DiscordTokenExpiration TIME NULL
	)`
	userSelectPassword     = `SELECT Password FROM users WHERE Username = $1`
	userSelectByUsername   = `SELECT ID, Username, Email, DiscordUID, DiscordAuthToken, DiscordTokenExpiration FROM users WHERE Username = $1 AND Password = $2`
	userSelectByDiscordUID = `SELECT ID, Username, Email, DiscordUID, DiscordAuthToken, DiscordTokenExpiration FROM users WHERE DiscordUID = $1`
	userSelectByID         = `SELECT ID, Username, Email, DiscordUID, DiscordAuthToken, DiscordTokenExpiration FROM users WHERE ID = $1`
	userInsert             = `INSERT INTO users (Username, Password, Email, DiscordUID, DiscordAuthToken, DiscordRefreshToken, DiscordTokenExpiration) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING ID`
)

type userTable struct {
	db *sql.DB
}

type User struct {
	ID                     string
	Username               string
	Email                  sql.NullString
	DiscordUID             sql.NullString
	DiscordAuthToken       sql.NullString
	DiscordTokenExpiration pq.NullTime
}

func (t *userTable) createTable() {
	_, err := t.db.Exec(userCreateTable)
	if err != nil {
		log.Fatalln(err)
	}
}

func (t *userTable) SelectByUsername(username string, password string) (*User, error) {
	if password == "" {
		return nil, fmt.Errorf("Password cannot be empty")
	}
	var passwordHash string
	row := t.db.QueryRow(userSelectPassword, username)
	if e := row.Scan(&passwordHash); e != nil {
		log.Println("User not found - ", e)
		return nil, e
	}
	pBytes, err := base64.StdEncoding.DecodeString(passwordHash)
	if err != nil {
		log.Println("Cannot decode password hash for user "+username+" - ", err)
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword(pBytes, []byte(password))
	if err != nil {
		log.Println("Wrong password for user "+username+" - ", err)
		return nil, err
	}
	user := &User{}
	row = t.db.QueryRow(userSelectByUsername, username, passwordHash)
	if e := row.Scan(&user.ID, &user.Username, &user.Email, &user.DiscordUID, &user.DiscordAuthToken, &user.DiscordTokenExpiration); e != nil {
		log.Println("User not found - ", e)
		return nil, e
	}
	return user, nil
}

func (t *userTable) SelectByDiscordUID(discordUID string) (*User, error) {
	user := &User{}
	row := t.db.QueryRow(userSelectByDiscordUID, discordUID)
	if e := row.Scan(&user.ID, &user.Username, &user.Email, &user.DiscordUID, &user.DiscordAuthToken, &user.DiscordTokenExpiration); e != nil {
		log.Println("User not found - ", e)
		return nil, e
	}
	return user, nil
}

func (t *userTable) SelectByID(userID string) (*User, error) {
	user := &User{}
	row := t.db.QueryRow(userSelectByID, userID)
	if e := row.Scan(&user.ID, &user.Username, &user.Email, &user.DiscordUID, &user.DiscordAuthToken, &user.DiscordTokenExpiration); e != nil {
		log.Println("User not found - ", e)
		return nil, e
	}
	return user, nil
}

func (t *userTable) InsertUser(user *User, password sql.NullString, discordRefreshToken sql.NullString) (*User, error) {
	var hashedPassword sql.NullString
	if password.Valid {
		pBytes := []byte(password.String)
		hash, err := bcrypt.GenerateFromPassword(pBytes, bcrypt.DefaultCost)
		if err != nil {
			log.Println("Cannot hash password for user " + user.Username)
			return nil, err
		}
		hashedPassword = sql.NullString{String: base64.StdEncoding.EncodeToString(hash), Valid: true}
	}
	row := t.db.QueryRow(userInsert, user.Username, hashedPassword, user.Email, user.DiscordUID, user.DiscordAuthToken, discordRefreshToken, user.DiscordTokenExpiration)
	if e := row.Scan(&user.ID); e != nil {
		log.Println("Cannot retrieve newly inserted ID")
		return nil, e
	}
	return user, nil
}
