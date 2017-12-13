package db

import (
	"log"
)

// Raffle type enum
type RaffleType int

// unfortunately these have to be pretty specific until I can come up with a better way to store them. or a more generic raffle system
const (
	_ RaffleType = iota
	JustRaffle
	// Raffles for the made in abyss server
	RaffleMIA
)

type RaffleEntry struct {
	id               int
	GuildUid         string
	UserUid          string
	RaffleType       RaffleType
	TicketCount      int
	RaffleData       string
	LastTicketUpdate int64
}

const (
	raffleQuery = `SELECT id, GuildUid, UserUid, RaffleType, TicketCount, RaffleData, LastTicketUpdate
					FROM raffle_entry AS re
					WHERE re.UserUid = $1 AND re.GuildUid = $2`

	raffleTable = `CREATE TABLE IF NOT EXISTS raffle_entry(
					id SERIAL NOT NULL PRIMARY KEY,
					GuildUid VARCHAR(20) NOT NULL,
					UserUid VARCHAR(20) NOT NULL,
					RaffleType SMALLINT NOT NULL,
					TicketCount INTEGER NOT NULL DEFAULT 0,
					RaffleData VARCHAR(1000) NOT NULL,
					LastTicketUpdate BIGINT NOT NULL DEFAULT 0,
					UNIQUE (GuildUid, UserUid)
				)`

	raffleInsert = `INSERT INTO raffle_entry (GuildUid, UserUid, RaffleType, TicketCount, RaffleData) VALUES ($1, $2, $3, $4, $5)`

	raffleUpdate = `UPDATE raffle_entry SET RaffleData = $2, TicketCount = TicketCount + $3, LastTicketUpdate = $4 WHERE id = $1`

	RaffleDataSeparator = "|"
)

func RaffleEntryAdd(entry RaffleEntry) error {
	_, err := moeDb.Exec(raffleInsert, entry.GuildUid, entry.UserUid, entry.RaffleType, entry.TicketCount, entry.RaffleData)
	if err != nil {
		log.Println("Error adding raffle entry to database, ", err)
		return err
	}
	return nil
}

func RaffleEntryUpdate(entry RaffleEntry, ticketAdd int) error {
	_, err := moeDb.Exec(raffleUpdate, entry.id, entry.RaffleData, ticketAdd, entry.LastTicketUpdate)
	if err != nil {
		log.Println("Error updating raffle entry to database, ", err)
		return err
	}
	return nil
}

func RaffleEntryQuery(userUid string, guildUid string) (raffleEntries []RaffleEntry, err error) {
	rows, err := moeDb.Query(raffleQuery, userUid, guildUid)
	if err != nil {
		log.Println("Error querying for raffle entries")
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var re RaffleEntry
		if err := rows.Scan(&re.id, &re.GuildUid, &re.UserUid, &re.RaffleType, &re.TicketCount, &re.RaffleData, &re.LastTicketUpdate); err != nil {
			log.Println("Error scanning raffle entry to object - ", err)
			return nil, err
		}
		raffleEntries = append(raffleEntries, re)
	}
	return raffleEntries, nil
}

func (re *RaffleEntry) SetRaffleData(raffleData string) {
	re.RaffleData = raffleData
}
