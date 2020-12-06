package db

import (
	"context"
	"log"

	"github.com/volatiletech/sqlboiler/boil"

	"github.com/volatiletech/sqlboiler/queries"

	"github.com/camd67/moebot/moebot_bot/util/db/models"
)

func PollQuery(id int) (*models.Poll, error) {
	result, err := models.FindPoll(context.Background(), moeDb, id)
	if err != nil {
		log.Println("Error querying for poll", err)
		return nil, err
	}
	result.R.PollOptions, err = PollOptionQuery(result.ID)
	if err != nil {
		log.Println("Error retreiving poll options", err)
		return nil, err
	}
	return result, nil
}

func PollsOpenQuery() (models.PollSlice, error) {
	result, err := models.Polls().All(context.Background(), moeDb)
	if err != nil {
		log.Println("Error querying for polls", err)
		return nil, err
	}
	return result, nil
}

func PollClose(id int) error {
	_, err := queries.Raw("UPDATE poll SET Open = FALSE WHERE id = $1", id).ExecContext(context.Background(), moeDb)
	if err != nil {
		log.Println("Error closing the poll", err)
		return err
	}
	return nil
}

func PollAdd(poll *models.Poll) error {
	err := poll.Insert(context.Background(), moeDb, boil.Whitelist("title", "channel_id", "user_uid"))
	if err != nil {
		log.Println("Error creating the poll", err)
		return err
	}
	return nil
}

func PollSetMessageId(poll *models.Poll) error {
	_, err := queries.Raw("UPDATE poll SET  message_uid = $1 WHERE id = $2",
		poll.MessageUID.String, poll.ID).ExecContext(context.Background(), moeDb)
	if err != nil {
		log.Println("Error updating the message Id for the poll", err)
		return err
	}
	return nil
}

func PollSetOptions(poll *models.Poll, options models.PollOptionSlice) *models.Poll {
	// adding options in db package since we need db handle even though we aren't adding them to db
	poll.SetPollOptions(context.Background(), moeDb, false, options...)
	return poll
}
