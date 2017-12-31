package db

import "log"

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

	customRoleQueryServer = `SELECT cr.id, cr.Trigger, r.RoleUid
		FROM customRole AS cr, server, role AS r
		WHERE server.GuildUid = $1`
	customRoleQueryTrigger = `SELECT r.RoleUid
		FROM customRole AS cr, server, role AS r
		WHERE server.GuildUid = $1 AND cr.trigger = $2`
	customRoleInsert = `INSERT INTO customRole(GuildId, RoleId, Trigger) VALUES ($1, $2, $3)`
)

func CustomRoleQuery(trigger string, guildUid string) (rId string, err error) {
	err = moeDb.QueryRow(customRoleQueryTrigger, guildUid, trigger).Scan(&rId)
	if err != nil {
		log.Println("Error querying for custom role {trigger, guildId} ", trigger, guildUid)
	}
	return
}

func CustomRoleAdd(trigger string, guildId int, roleId int) (err error) {
	_, err = moeDb.Exec(customRoleInsert, guildId, roleId, trigger)
	if err != nil {
		log.Println("Error inserting new custom role: {trigger, guildId, roleId}", trigger, guildId, roleId)
	}
	return
}
