package db

import (
	"database/sql"
	"log"
	"time"
)

const (
	guildEventCreateTable = `CREATE TABLE IF NOT EXISTS guildEvent(
		ID SERIAL NOT NULL PRIMARY KEY,
		Description VARCHAR NOT NULL,
		Date TIME NOT NULL,
		Points INTEGER NOT NULL DEFAULT 0,
		GuildUID VARCHAR(20) NOT NULL
	)`
	guildEventSelectByRange = `SELECT ID, Description, Date, Points, GuildUID FROM guildEvent WHERE Date >= $1 AND Date < $2 AND GuildUID = $3`
	guildEventInsert        = `INSERT INTO guildEvent (Description, Date, Points, GuildUID) VALUES ($1, $2, $3, $4) RETURNING ID`
)

var guildEventUpdateTable = []string{
	"ALTER TABLE guildEvent ADD COLUMN IF NOT EXISTS Points INTEGER NOT NULL DEFAULT 0",
}

type GuildEvent struct {
	ID          int          `json:"id"`
	Description string       `json:"description"`
	Date        time.Time    `json:"date"`
	Points      int          `json:"points"`
	GuildUID    string       `json:"guilduid"`
	Users       []*EventUser `json:"users"`
}

type guildEventTable struct {
	db *sql.DB
}

func (t *guildEventTable) createTable() {
	_, err := t.db.Exec(guildEventCreateTable)
	if err != nil {
		log.Fatalln(err)
	}
	for _, u := range guildEventUpdateTable {
		_, err = t.db.Exec(u)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func (t *guildEventTable) SelectEvents(startDate time.Time, endDate time.Time, guildUID string) ([]*GuildEvent, error) {
	events := []*GuildEvent{}
	rows, err := t.db.Query(guildEventSelectByRange, startDate, endDate, guildUID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		event := &GuildEvent{}
		if err = rows.Scan(&event.ID, &event.Description, &event.Date, &event.Points, &event.GuildUID); err != nil {
			log.Println("Error scanning event - ", err)
			return nil, err
		}
		eus, err := t.getEventUsers(event.ID)
		if err != nil {
			return nil, err
		}
		event.Users = eus
		events = append(events, event)
	}
	return events, nil
}

func (t *guildEventTable) InsertEvent(event *GuildEvent) (*GuildEvent, error) {
	row := t.db.QueryRow(guildEventInsert, event.Description, event.Date, event.Points, event.GuildUID)
	if err := row.Scan(&event.ID); err != nil {
		log.Println("Cannot insert new event - ", err)
		return nil, err
	}
	euTable := &eventUserTable{t.db}
	for _, u := range event.Users {
		_, err := euTable.InsertUser(u.UserUID, event.ID)
		if err != nil {
			return nil, err
		}
	}
	return event, nil
}

func (t *guildEventTable) getEventUsers(eventID int) ([]*EventUser, error) {
	table := &eventUserTable{t.db}
	return table.SelectUsers(eventID)
}
