package db

import "log"

type poll struct {
	Id        int
	Options   []*pollOption
	Title     string
	Open      bool
	UserId    string
	ChannelId string
}

const (
	pollTable = `CREATE TABLE IF NOT EXISTS poll(
		Id SERIAL NOT NULL PRIMARY KEY,
		Title VARCHAR(100) NOT NULL,
		ChannelId VARCHAR(20) NOT NULL,
		UserId VARCHAR(20) NOT NULL,
		Open BOOLEAN NOT NULL DEFAULT 1
	)`

	pollSelect = `SELECT (Id, Title, ChannelId, UserId, Open) FROM poll WHERE Id = $1`

	pollSelectOpen = `SELECT (Id, Title, ChannelId, UserId, Open) FROM poll WHERE Open = 1`
)

func PollQuery(id int) (*poll, error) {
	var err error
	row := moeDb.QueryRow(pollSelect, id)
	result := new(poll)
	if err = row.Scan(&result.Id, &result.Title, &result.ChannelId, &result.UserId, &result.Open); err != nil {
		log.Println("Error querying for poll", err)
		return nil, err
	}
	result.Options, err = PollOptionQuery(result.Id)
	return result, nil
}

func PollsOpenQuery() ([]*poll, error) {
	rows, err := moeDb.Query(pollSelectOpen, nil)
	if err != nil {
		log.Println("Error querying for polls", err)
		return nil, err
	}
	result := []*poll{}
	for rows.Next() {
		p := new(poll)
		rows.Scan(&p.Id, &p.Title, &p.ChannelId, &p.UserId, &p.Open)
		result = append(result, p)
	}
	return result, nil
}
