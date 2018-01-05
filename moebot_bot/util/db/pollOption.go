package db

import "log"

type pollOption struct {
	Id          int
	PollId      int
	Number      int
	Description string
	Votes       int
}

const (
	pollOptionTable = `CREATE TABLE IF NOT EXISTS poll_option(
		Id SERIAL NOT NULL PRIMARY KEY,
		PollId INTEGER REFERENCES poll(Id) ON DELETE CASCADE,
		Number INTEGER NOT NULL,
		Description VARCHAR(200) NOT NULL,
		Votes INTEGER NOT NULL DEFAULT 0
	)`

	pollOptionSelectPoll = `SELECT (Id, PollId, Number, Description, Votes) FROM poll_option WHERE PollId = $1 ORDER BY Number`
)

func PollOptionQuery(pollId int) ([]*pollOption, error) {
	rows, err := moeDb.Query(pollOptionSelectPoll, pollId)
	if err != nil {
		log.Println("Error querying for poll options", err)
		return nil, err
	}
	result := []*pollOption{}
	for rows.Next() {
		option := new(pollOption)
		rows.Scan(&option.Id, &option.PollId, &option.Number, &option.Description, &option.Votes)
		result = append(result, option)
	}
	return result, nil
}
