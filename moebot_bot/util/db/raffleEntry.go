package db

import (
	"context"
	"log"

	"github.com/volatiletech/sqlboiler/queries/qm"

	"github.com/camd67/moebot/moebot_bot/util/db/models"
	"github.com/volatiletech/sqlboiler/boil"

	"github.com/camd67/moebot/moebot_bot/util/db/types"
)

// unfortunately these have to be pretty specific until I can come up with a better way to store them. or a more generic raffle system
const (
	_ types.RaffleType = iota
	JustRaffle
	// Raffles for the made in abyss server
	RaffleMIA
)

const (
	raffleSelect = `SELECT Id, GuildUid, UserUid, RaffleType, TicketCount, RaffleData, LastTicketUpdate `

	raffleQuery = raffleSelect + `FROM raffle_entry AS re
					WHERE re.UserUid = $1 AND re.GuildUid = $2`

	raffleQueryAny = raffleSelect + `FROM raffle_entry AS re
						WHERE re.GuildUid = $1`

	raffleTable = `CREATE TABLE IF NOT EXISTS raffle_entry(
					Id SERIAL NOT NULL PRIMARY KEY,
					GuildUid VARCHAR(20) NOT NULL,
					UserUid VARCHAR(20) NOT NULL,
					RaffleType SMALLINT NOT NULL,
					TicketCount INTEGER NOT NULL DEFAULT 0,
					RaffleData VARCHAR(1000) NOT NULL,
					LastTicketUpdate BIGINT NOT NULL DEFAULT 0,
					UNIQUE (GuildUid, UserUid)
				)`

	raffleInsert = `INSERT INTO raffle_entry (GuildUid, UserUid, RaffleType, TicketCount, RaffleData) VALUES ($1, $2, $3, $4, $5)`

	raffleUpdate = `UPDATE raffle_entry SET RaffleData = $2, TicketCount = TicketCount + $3, LastTicketUpdate = $4 WHERE Id = $1`

	raffleUpdateMany = `UPDATE raffle_entry SET TicketCount = TicketCount + $1 WHERE Id = ANY ($2::integer[])`

	RaffleDataSeparator = "|"
)

func RaffleEntryAdd(entry *models.RaffleEntry) error {
	err := entry.Insert(context.Background(), moeDb, boil.Infer())
	if err != nil {
		log.Println("Error adding raffle entry to database, ", err)
		return err
	}
	return nil
}

func RaffleEntryUpdate(entry *models.RaffleEntry, ticketAdd int) error {
	entry.TicketCount += ticketAdd
	_, err := entry.Update(context.Background(), moeDb, boil.Infer())
	if err != nil {
		log.Println("Error updating raffle entry to database, ", err)
		return err
	}
	return nil
}

func RaffleEntryUpdateMany(entries models.RaffleEntrySlice, ticketAdd int) error {
	// for _, entry := range entries {
	// 	entry.TicketCount += ticketAdd
	// 	err := entry.Update(context.Background(), moeDb, boil.Whitelist("ticket_count"))
	// 	if err != nil {
	// 		log.Println("Error updating raffle entry to database, ", err)
	// 		return err
	// 	}
	// }
	log.Println("This wasn't implemented in the sqlboiler update.")
	return nil
}

func RaffleEntryQuery(userUid string, guildUid string) (raffleEntries models.RaffleEntrySlice, err error) {
	raffleEntries, err = models.RaffleEntries(qm.Where("user_uid = ? AND guild_uid = ?", userUid, guildUid)).All(context.Background(), moeDb)
	if err != nil {
		log.Println("Error querying for raffle entries by user")
		return nil, err
	}
	return raffleEntries, nil
}

func RaffleEntryQueryAny(guildUid string) (raffleEntries models.RaffleEntrySlice, err error) {
	raffleEntries, err = models.RaffleEntries(qm.Where("guild_uid = ?", guildUid)).All(context.Background(), moeDb)
	if err != nil {
		log.Println("Error querying for raffle entries by guild")
		return nil, err
	}
	return raffleEntries, nil
}
