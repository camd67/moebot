package db

import "log"

type Poll struct {
	Id         int
	Options    []*PollOption
	Title      string
	Open       bool
	ChannelId  int
	UserUid    string
	MessageUid string
}

const (
	pollTable = `CREATE TABLE IF NOT EXISTS poll(
		Id SERIAL NOT NULL PRIMARY KEY,
		Title VARCHAR(100) NOT NULL,
		ChannelId INTEGER NOT NULL REFERENCES channel(Id) ON DELETE CASCADE,
		UserUid VARCHAR(20) NOT NULL,
		MessageUid VARCHAR(20),
		Open BOOLEAN NOT NULL DEFAULT TRUE
	)`

	pollSelect = `SELECT Id, Title, ChannelId, UserUid, MessageUid, Open FROM poll WHERE Id = $1`

	pollSelectOpen = `SELECT Id, Title, ChannelId, UserUid, MessageUid, Open FROM poll WHERE Open = TRUE`

	pollClose = `UPDATE poll SET Open = FALSE WHERE Id = $1`

	pollInsert = `INSERT INTO poll (Title, ChannelId, UserUid) VALUES($1, $2, $3) RETURNING Id`

	pollSetMessageId = `UPDATE poll SET MessageUid = $1 WHERE Id = $2`
)

func PollQuery(id int) (*Poll, error) {
	var err error
	row := moeDb.QueryRow(pollSelect, id)
	result := new(Poll)
	if err = row.Scan(&result.Id, &result.Title, &result.ChannelId, &result.UserUid, &result.MessageUid, &result.Open); err != nil {
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
		rows.Scan(&p.Id, &p.Title, &p.ChannelId, &p.UserUid, &p.MessageUid, &p.Open)
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
	err := moeDb.QueryRow(pollInsert, poll.Title, poll.ChannelId, poll.UserUid).Scan(&poll.Id)
	if err != nil {
		log.Println("Error creating the poll", err)
		return err
	}
	return nil
}

func PollSetMessageId(poll *Poll) error {
	_, err := moeDb.Exec(pollSetMessageId, poll.MessageUid, poll.Id)
	if err != nil {
		log.Println("Error updating the message Id for the poll", err)
		return err
	}
	return nil
}
