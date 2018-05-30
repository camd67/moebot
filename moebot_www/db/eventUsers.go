package db

import (
	"database/sql"
	"log"
)

const (
	eventUserCreateTable = `CREATE TABLE IF NOT EXISTS eventUser(
		UserUID VARCHAR(20) NOT NULL,
		EventID INTEGER NOT NULL REFERENCES guildEvent(Id) ON DELETE CASCADE
		PRIMARY KEY(UserUID, EventID)
	)`
	eventUserSelect = `SELECT UserUID FROM eventUser WHERE EventID = $1`
	eventUserInsert = `INSERT INTO eventUser (UserUID, EventID) VALUES ($1, $2)`
)

type EventUser struct {
	UserUID string `json:"useruid"`
}

type eventUserTable struct {
	db *sql.DB
}

func (t *eventUserTable) createTable() {
	_, err := t.db.Exec(eventUserCreateTable)
	if err != nil {
		log.Fatalln(err)
	}
}

func (t *eventUserTable) SelectUsers(eventId int) ([]*EventUser, error) {
	rows, err := t.db.Query(eventUserSelect, eventId)
	if err != nil {
		log.Println("Error querying event users - ", err)
		return nil, err
	}
	users := []*EventUser{}
	for rows.Next() {
		eu := &EventUser{}
		err = rows.Scan(&eu.UserUID)
		if err != nil {
			log.Println("Error querying event users - ", err)
			return nil, err
		}
		users = append(users, eu)
	}
	return users, nil
}

func (t *eventUserTable) InsertUser(userUID string, eventID int) (*EventUser, error) {
	_, err := t.db.Exec(eventUserInsert, userUID, eventID)
	if err != nil {
		log.Println("Error inserting event user - ", err)
		return nil, err
	}
	return &EventUser{userUID}, nil
}
