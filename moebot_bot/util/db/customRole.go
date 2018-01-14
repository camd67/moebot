package db

import (
	"database/sql"
	"log"
	"strings"
)

type CustomRole struct {
	Id      int
	GuildId int
	RoleId  int
	Trigger string
}

const (
	customRoleTable = `CREATE TABLE IF NOT EXISTS customRole(
		Id SERIAL NOT NULL PRIMARY KEY,
		GuildId INTEGER REFERENCES server(id) ON DELETE CASCADE,
		RoleId INTEGER REFERENCES role(id) ON DELETE CASCADE,
		Trigger VARCHAR(50) NOT NULL
	)`

	customRoleQueryServer = `SELECT cr.Trigger
		FROM customRole AS cr
		JOIN server ON server.id = cr.GuildId
		WHERE server.GuildUid = $1`
	customRoleQueryTrigger = `SELECT r.RoleUid
		FROM customRole AS cr
		JOIN server AS s ON cr.GuildId = s.Id
		JOIN role AS r ON cr.RoleId = r.Id
		WHERE s.GuildUid = $1 AND cr.trigger = $2`
	customRoleInsert = `INSERT INTO customRole(GuildId, RoleId, Trigger) VALUES ($1, $2, $3)`
	customRoleSearch = `SELECT customRole.Id FROM customRole, server WHERE Trigger = $1 AND GuildUid = $2`
	customRoleDelete = `DELETE FROM customRole WHERE Id = $1`
)

func CustomRoleRowExists(trigger string, guildUid string) (id int, exists bool) {
	row := moeDb.QueryRow(customRoleSearch, trigger, guildUid)
	err := row.Scan(&id)
	if err == sql.ErrNoRows {
		return -1, false
	}
	return id, true
}

func CustomRoleQuery(trigger string, guildUid string) (rId string, err error) {
	err = moeDb.QueryRow(customRoleQueryTrigger, guildUid, trigger).Scan(&rId)
	if err != nil {
		log.Println("Error querying for custom role {trigger, guildId} ", trigger, guildUid)
	}
	return
}

func CustomRoleQueryServer(guildUid string) (triggers []string, err error) {
	rows, err := moeDb.Query(customRoleQueryServer, guildUid)
	if err != nil {
		log.Println("Error querying for server:", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var trigger string
		if err = rows.Scan(&trigger); err != nil {
			log.Println("Error scanning from customRole table:", err)
			return
		}
		triggers = append(triggers, trigger)
	}
	return
}

func CustomRoleAdd(trigger string, guildId int, roleId int) (err error) {
	_, err = moeDb.Exec(customRoleInsert, guildId, roleId, strings.TrimSpace(trigger))
	if err != nil {
		log.Println("Error inserting new custom role: {trigger, guildId, roleId}", trigger, guildId, roleId)
	}
	return
}

func CustomRoleDelete(id int) int64 {
	r, err := moeDb.Exec(customRoleDelete, id)
	if err != nil {
		log.Println("Error deleting custom role: ", id)
		return -1
	}
	rCount, _ := r.RowsAffected()
	return rCount
}
