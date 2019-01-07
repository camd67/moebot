package db

import (
	"context"
	"log"

	"github.com/volatiletech/sqlboiler/boil"

	"github.com/volatiletech/sqlboiler/queries/qm"

	"github.com/camd67/moebot/moebot_bot/util/db/models"
)

func PollOptionQuery(pollId int) (models.PollOptionSlice, error) {
	options, err := models.PollOptions(qm.Where("poll_id = ?", pollId), qm.OrderBy("id")).All(context.Background(), moeDb)
	if err != nil {
		log.Println("Error querying for poll options", err)
		return nil, err
	}
	return options, nil
}

func PollOptionAdd(poll *models.Poll) error {
	for _, o := range poll.R.PollOptions {
		err := o.Insert(context.Background(), moeDb, boil.Whitelist("poll_id", "reaction_id", "reaction_name", "description"))
		if err != nil {
			log.Println("Error inserting poll options", err)
			return err
		}
	}
	return nil
}

func PollOptionUpdateVotes(poll *models.Poll) error {
	for _, o := range poll.R.PollOptions {
		_, err := o.Update(context.Background(), moeDb, boil.Whitelist("votes"))
		if err != nil {
			log.Println("Error updating poll votes", err)
			return err
		}
	}
	return nil
}
