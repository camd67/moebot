package db

import (
	"database/sql"
	"log"
)

type Nomination struct {
	ID        int
	Entries   []*NominationEntry
	Title     string
	Open      bool
	ChannelID int
}

const (
	nominationTable = `CREATE TABLE IF NOT EXISTS nomination(
		Id SERIAL NOT NULL PRIMARY KEY,
		Title VARCHAR(100) NOT NULL,
		ChannelId INTEGER NOT NULL REFERENCES channel(Id) ON DELETE CASCADE,
		Open BOOLEAN NOT NULL DEFAULT TRUE
	)`

	nominationSelectByTitleAndChannelID = `SELECT nomination.Id, nomination.Title, nomination.ChannelId, nomination.Open WHERE ChannelId = $1 AND Title = $2`

	nominationClose = `UPDATE nomination SET Open = FALSE WHERE Id = $1`

	nominationInsert = `INSERT INTO nomination (Title, ChannelId) VALUES($1, $2) RETURNING Id`
)

func NominationQuery(title string, channelID int) (*Nomination, error) {
	var err error
	row := moeDb.QueryRow(nominationSelectByTitleAndChannelID, title, channelID)
	result := new(Nomination)
	if err = row.Scan(&result.ID, &result.Title, &result.ChannelID, &result.Open); err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error querying for nomination", err)
		}
		return nil, err
	}
	result.Entries, err = NominationOptionsQuery(result.ID)
	if err != nil {
		log.Println("Error retrieving nomination entries", err)
		return nil, err
	}
	return result, nil
}

func NominationOpen(title string, channelID int) (*Nomination, error) {
	var err error
	result := &Nomination{
		Entries:   []*NominationEntry{},
		Title:     title,
		ChannelID: channelID,
		Open:      true,
	}
	err = moeDb.QueryRow(nominationInsert, title, channelID).Scan(&result.ID)
	if err != nil {
		log.Println("Error creating the nomination", err)
		return nil, err
	}
	return result, nil
}

func NominationClose(title string, channelID int) (*Nomination, error) {
	result, err := NominationQuery(title, channelID)
	if err != nil {
		return nil, err
	}
	_, err = moeDb.Exec(nominationClose, result.ID)
	if err != nil {
		log.Println("Error closing the nomination", err)
		return nil, err
	}
	result.Open = false
	return result, nil
}
