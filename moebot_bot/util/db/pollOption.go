package db

import "log"

type PollOption struct {
	Id           int
	PollId       int
	ReactionId   string
	ReactionName string
	Description  string
	Votes        int
}

const (
	pollOptionTable = `CREATE TABLE IF NOT EXISTS poll_option(
		Id SERIAL NOT NULL PRIMARY KEY,
		PollId INTEGER REFERENCES poll(Id) ON DELETE CASCADE,
		ReactionId VARCHAR(20) NOT NULL,
		ReactionName VARCHAR(30) NOT NULL,
		Description VARCHAR(200) NOT NULL,
		Votes INTEGER NOT NULL DEFAULT 0
	)`

	pollOptionSelectPoll = `SELECT Id, PollId, ReactionId, ReactionName, Description, Votes FROM poll_option WHERE PollId = $1 ORDER BY Id`

	pollOptionInsert = `INSERT INTO poll_option (PollId, ReactionId, ReactionName, Description) VALUES($1, $2, $3, $4) RETURNING Id`

	pollOptionUpdateVotes = `UPDATE poll_option SET Votes = $1 WHERE Id = $2`
)

func PollOptionQuery(pollId int) ([]*PollOption, error) {
	rows, err := moeDb.Query(pollOptionSelectPoll, pollId)
	if err != nil {
		log.Println("Error querying for poll options", err)
		return nil, err
	}
	result := []*PollOption{}
	for rows.Next() {
		option := new(PollOption)
		err = rows.Scan(&option.Id, &option.PollId, &option.ReactionId, &option.ReactionName, &option.Description, &option.Votes)
		if err != nil {
			log.Println("Error querying for poll options", err)
			return nil, err
		}
		result = append(result, option)
	}
	return result, nil
}

func PollOptionAdd(poll *Poll) error {
	for _, o := range poll.Options {
		err := moeDb.QueryRow(pollOptionInsert, poll.Id, o.ReactionId, o.ReactionName, o.Description).Scan(&o.Id)
		if err != nil {
			log.Println("Error inserting poll options", err)
			return err
		}
	}
	return nil
}

func PollOptionUpdateVotes(poll *Poll) error {
	for _, o := range poll.Options {
		_, err := moeDb.Exec(pollOptionUpdateVotes, o.Votes, o.Id)
		if err != nil {
			log.Println("Error updating poll votes", err)
			return err
		}
	}
	return nil
}
