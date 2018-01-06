package db

import "log"

type Poll struct {
	Id        int
	Options   []*PollOption
	Title     string
	Open      bool
	UserId    string
	ChannelId string
	MessageId string
}

const (
	pollTable = `CREATE TABLE IF NOT EXISTS poll(
		Id SERIAL NOT NULL PRIMARY KEY,
		Title VARCHAR(100) NOT NULL,
		ChannelId VARCHAR(20) NOT NULL,
		UserId VARCHAR(20) NOT NULL,
		MessageId VARCHAR(20),
		Open BOOLEAN NOT NULL DEFAULT TRUE
	)`

	pollSelect = `SELECT Id, Title, ChannelId, UserId, MessageId, Open FROM poll WHERE Id = $1`

	pollSelectOpen = `SELECT Id, Title, ChannelId, UserId, MessageId, Open FROM poll WHERE Open = TRUE`

	pollClose = `UPDATE poll SET Open = FALSE WHERE Id = $1`

	pollInsert = `INSERT INTO poll (Title, ChannelId, UserId) VALUES($1, $2, $3) RETURNING Id`

	pollSetMessageId = `UPDATE poll SET MessageId = $1 WHERE Id = $2`
)

func PollQuery(id int) (*Poll, error) {
	var err error
	row := moeDb.QueryRow(pollSelect, id)
	result := new(Poll)
	if err = row.Scan(&result.Id, &result.Title, &result.ChannelId, &result.UserId, &result.MessageId, &result.Open); err != nil {
		log.Println("Error querying for poll", err)
		return nil, err
	}
	result.Options, err = PollOptionQuery(result.Id)
	if err != nil {
		log.Println("Error retreiving poll options", err)
		return nil, err
	}
	return result, nil
}

func PollsOpenQuery() ([]*Poll, error) {
	rows, err := moeDb.Query(pollSelectOpen)
	if err != nil {
		log.Println("Error querying for polls", err)
		return nil, err
	}
	result := []*Poll{}
	for rows.Next() {
		p := new(Poll)
		rows.Scan(&p.Id, &p.Title, &p.ChannelId, &p.UserId, &p.MessageId, &p.Open)
		result = append(result, p)
	}
	return result, nil
}

func PollClose(id int) error {
	_, err := moeDb.Exec(pollClose, id)
	if err != nil {
		log.Println("Error closing the poll", err)
		return err
	}
	return nil
}

func PollAdd(poll *Poll) error {
	err := moeDb.QueryRow(pollInsert, poll.Title, poll.ChannelId, poll.UserId).Scan(&poll.Id)
	if err != nil {
		log.Println("Error creating the poll", err)
		return err
	}
	return nil
}

func PollSetMessageId(poll *Poll) error {
	_, err := moeDb.Exec(pollSetMessageId, poll.MessageId, poll.Id)
	if err != nil {
		log.Println("Error updating the message Id for the poll", err)
		return err
	}
	return nil
}
