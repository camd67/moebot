package db

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
)

// Permission enum
type Permission int

const (
	// Default permission level, no permissions regarding what can or can't be done
	PermAll Permission = 2
	// Mod level permission, allowed to do some server changing commands
	PermMod Permission = 50
	// Guild Owner permission. Essentially a master
	PermGuildOwner Permission = 90
	// Used to disable something, no one can have this permission level
	PermNone Permission = 100
	// Master level permission, can't ever be ignored or disabled
	PermMaster Permission = 101
)

type Role struct {
	Id                         int
	ServerId                   int
	GroupId                    int
	RoleUid                    string
	Permission                 Permission
	ConfirmationMessage        sql.NullString
	ConfirmationSecurityAnswer sql.NullString
	Trigger                    sql.NullString
}

const (
	roleTable = `CREATE TABLE IF NOT EXISTS role(
		Id SERIAL NOT NULL PRIMARY KEY,
		ServerId INTEGER REFERENCES server(Id) ON DELETE CASCADE,
		GroupId INTEGER REFERENCES role_group(Id) ON DELETE CASCADE,
		RoleUid VARCHAR(20) NOT NULL UNIQUE,
		Permission SMALLINT NOT NULL DEFAULT 2,
		ConfirmationMessage VARCHAR CONSTRAINT role_confirmation_message_length CHECK (char_length(ConfirmationMessage) <= 1900),
		ConfirmationSecurityAnswer VARCHAR CONSTRAINT role_confirmation_security_answer_length CHECK (char_length(ConfirmationMessage) <= 1900),
		Trigger TEXT CONSTRAINT role_trigger_length CHECK(char_length(Trigger) <= 100)
	)`

	RoleMaxTriggerLength       = 100
	RoleMaxTriggerLengthString = "100"

	roleQueryServerRole  = `SELECT Id, ServerId, GroupId, RoleUid, Permission, ConfirmationMessage, ConfirmationSecurityAnswer, Trigger FROM role WHERE RoleUid = $1 AND ServerId = $2`
	roleQueryServer      = `SELECT Id, ServerId, GroupId, RoleUid, Permission, ConfirmationMessage, ConfirmationSecurityAnswer, Trigger FROM role WHERE ServerId = $1`
	roleQuery            = `SELECT Id, ServerId, GroupId, RoleUid, Permission, ConfirmationMessage, ConfirmationSecurityAnswer, Trigger FROM role WHERE Id = $1`
	roleQueryTrigger     = `SELECT Id, ServerId, GroupId, RoleUid, Permission, ConfirmationMessage, ConfirmationSecurityAnswer, Trigger FROM role WHERE UPPER(Trigger) = UPPER($1) AND ServerId = $2`
	roleQueryGroup       = `SELECT Id, ServerId, GroupId, RoleUid, Permission, ConfirmationMessage, ConfirmationSecurityAnswer, Trigger FROM role WHERE GroupId = $1`
	roleQueryPermissions = `SELECT Permission FROM role WHERE RoleUid = ANY ($1::varchar[])`

	roleUpdate = `UPDATE role SET GroupId = $2, Permission = $3, ConfirmationMessage = $4, ConfirmationSecurityAnswer = $5, Trigger = $6 WHERE Id = $1`

	roleInsert = `INSERT INTO role(ServerId, RoleUid, GroupId, Permission, ConfirmationMessage, ConfirmationSecurityAnswer, Trigger) VALUES($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	roleDelete = `DELETE FROM role WHERE role.RoleUid = $1 AND role.ServerId = (SELECT server.id FROM server WHERE server.guilduid = $2)`
)

var (
	roleUpdateTable = []string{
		`ALTER TABLE role ADD COLUMN IF NOT EXISTS ConfirmationMessage VARCHAR`,
		`ALTER TABLE role ADD COLUMN IF NOT EXISTS ConfirmationSecurityAnswer VARCHAR`,
		`ALTER TABLE role DROP COLUMN IF EXISTS RoleType`,
		`ALTER TABLE role ADD COLUMN IF NOT EXISTS Trigger TEXT`,
		`ALTER TABLE role DROP CONSTRAINT IF EXISTS role_trigger_length`,
		`ALTER TABLE role ADD CONSTRAINT role_trigger_length CHECK(char_length(Trigger) <= 100)`,
		`ALTER TABLE role DROP CONSTRAINT IF EXISTS role_confirmation_message_length`,
		`ALTER TABLE role ADD CONSTRAINT role_confirmation_message_length CHECK(char_length(ConfirmationMessage) <= 1900)`,
		`ALTER TABLE role DROP CONSTRAINT IF EXISTS role_confirmation_security_answer_length`,
		`ALTER TABLE role ADD CONSTRAINT role_confirmation_security_answer_length CHECK(char_length(ConfirmationSecurityAnswer) <= 1900)`,
		`ALTER TABLE role ADD COLUMN IF NOT EXISTS GroupId INTEGER REFERENCES role_group(Id) ON DELETE CASCADE`,
		`ALTER TABLE role ALTER COLUMN Permission SET DEFAULT 2`,
	}
)

func RoleInsertOrUpdate(role Role) error {
	row := moeDb.QueryRow(roleQueryServerRole, role.RoleUid, role.ServerId)
	var r Role
	if err := row.Scan(&r.Id, &r.ServerId, &r.GroupId, &r.RoleUid, &r.Permission, &r.ConfirmationMessage, &r.ConfirmationSecurityAnswer, &r.Trigger); err != nil {
		if err == sql.ErrNoRows {
			// no row, so insert it add in default values
			if role.Permission == -1 {
				role.Permission = PermAll
			}
			_, err = moeDb.Exec(roleInsert, role.ServerId, strings.TrimSpace(role.RoleUid), role.GroupId, role.Permission, role.ConfirmationMessage,
				role.ConfirmationSecurityAnswer, role.Trigger)
			if err != nil {
				log.Println("Error inserting role to db")
				return err
			}
		} else {
			log.Println("Error scanning for role", err)
			return err
		}
	} else {
		// got a row, update it
		if role.Permission > 0 {
			r.Permission = role.Permission
		}
		if role.ConfirmationMessage.Valid {
			r.ConfirmationMessage = role.ConfirmationMessage
		}
		if role.ConfirmationSecurityAnswer.Valid {
			r.ConfirmationSecurityAnswer = role.ConfirmationSecurityAnswer
		}
		if role.Trigger.Valid {
			r.Trigger = role.Trigger
		}
		if role.GroupId > 0 {
			r.GroupId = role.GroupId
		}
		_, err = moeDb.Exec(roleUpdate, r.Id, r.GroupId, r.Permission, r.ConfirmationMessage, r.ConfirmationSecurityAnswer, r.Trigger)
		if err != nil {
			log.Println("Error updating role to db: Id " + strconv.Itoa(r.Id))
			return err
		}
	}
	return nil
}

func RoleQueryOrInsert(role Role) (r Role, err error) {
	row := moeDb.QueryRow(roleQueryServerRole, role.ServerId, role.RoleUid)
	if err = row.Scan(&r.Id, &r.ServerId, &r.GroupId, &r.RoleUid, &r.Permission, &r.ConfirmationMessage, &r.ConfirmationSecurityAnswer, &r.Trigger); err != nil {
		if err == sql.ErrNoRows {
			// no row, so insert it add in default values
			if role.Permission == -1 {
				role.Permission = PermAll
			}
			var insertId int
			err = moeDb.QueryRow(roleInsert, role.ServerId, strings.TrimSpace(role.RoleUid), role.GroupId, role.Permission, role.ConfirmationMessage,
				role.ConfirmationSecurityAnswer, role.Trigger).Scan(&insertId)
			if err != nil {
				log.Println("Error inserting role to db")
				return
			}
			row := moeDb.QueryRow(roleQuery, insertId)
			if err = row.Scan(&r.Id, &r.ServerId, &r.GroupId, &r.RoleUid, &r.Permission, &r.ConfirmationMessage, &r.ConfirmationSecurityAnswer, &r.Trigger); err != nil {
				log.Println("Failed to read the newly inserted Role row. This should pretty much never happen...", err)
				return Role{}, err
			}
		}
	}
	// got a row, return it
	return
}

func RoleQueryServer(s Server) (roles []Role, err error) {
	rows, err := moeDb.Query(roleQueryServer, s.Id)
	if err != nil {
		log.Println("Error querying for role", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var r Role
		if err = rows.Scan(&r.Id, &r.ServerId, &r.GroupId, &r.RoleUid, &r.Permission, &r.ConfirmationMessage, &r.ConfirmationSecurityAnswer,
			&r.Trigger); err != nil {

			log.Println("Error scanning from role table:", err)
			return
		}
		roles = append(roles, r)
	}
	return
}

func RoleQueryGroup(groupId int) (roles []Role, err error) {
	rows, err := moeDb.Query(roleQueryGroup, groupId)
	if err != nil {
		log.Println("Error querying for role", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var r Role
		if err = rows.Scan(&r.Id, &r.ServerId, &r.GroupId, &r.RoleUid, &r.Permission, &r.ConfirmationMessage, &r.ConfirmationSecurityAnswer,
			&r.Trigger); err != nil {

			log.Println("Error scanning from role table:", err)
			return
		}
		roles = append(roles, r)
	}
	return
}

func RoleQueryTrigger(trigger string, serverId int) (r Role, err error) {
	row := moeDb.QueryRow(roleQueryTrigger, trigger, serverId)
	err = row.Scan(&r.Id, &r.ServerId, &r.GroupId, &r.RoleUid, &r.Permission, &r.ConfirmationMessage, &r.ConfirmationSecurityAnswer, &r.Trigger)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error querying for role by trigger", err)
	}
	// return whatever we get, error or row
	return
}

func RoleQueryRoleUid(roleUid string, serverId int) (r Role, err error) {
	row := moeDb.QueryRow(roleQueryServerRole, roleUid, serverId)
	err = row.Scan(&r.Id, &r.ServerId, &r.GroupId, &r.RoleUid, &r.Permission, &r.ConfirmationMessage, &r.ConfirmationSecurityAnswer, &r.Trigger)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error querying for role by UID and serverID", err)
	}
	// return whatever we get, error or row
	return
}

func RoleQueryPermission(roleUids []string) (p []Permission) {
	idCollection := "{" + strings.Join(roleUids, ",") + "}"
	r, err := moeDb.Query(roleQueryPermissions, idCollection)
	if err != nil {
		log.Println("Error querying for user permissions", err)
		return
	}
	for r.Next() {
		var newPerm Permission
		r.Scan(&newPerm)
		p = append(p, newPerm)
	}
	return
}

func RoleDelete(roleUid string, guildUid string) error {
	_, err := moeDb.Exec(roleDelete, roleUid, guildUid)
	if err != nil {
		log.Println("Error deleting role: ", roleUid)
	}
	return err
}

func GetPermissionFromString(s string) Permission {
	toCheck := strings.ToUpper(s)
	if toCheck == "ALL" {
		return PermAll
	} else if toCheck == "MOD" {
		return PermMod
	} else if toCheck == "GUILD OWNER" || toCheck == "GO" {
		return PermGuildOwner
	} else if toCheck == "NONE" {
		return PermNone
	} else if toCheck == "MASTER" {
		return PermMaster
	} else {
		return -1
	}
}

func SprintPermission(p Permission) string {
	switch p {
	case PermAll:
		return "All"
	case PermMod:
		return "Mod"
	case PermGuildOwner:
		return "Guild Owner"
	case PermNone:
		return "None"
	case PermMaster:
		return "Master"
	default:
		return "Unknown"
	}
}

/**
Currently only a subset of roles are assignable by the bot
*/
func IsAssignablePermissionLevel(p Permission) bool {
	return p == PermMod || p == PermAll
}

/**
Gets a string representing all the possible assignable permission levels
*/
func GetAssignableRoles() string {
	return "{All, Mod}"
}

func roleCreateTable() {
	_, err := moeDb.Exec(roleTable)
	if err != nil {
		log.Println("Error creating role table", err)
		return
	}
	for _, alter := range roleUpdateTable {
		_, err = moeDb.Exec(alter)
		if err != nil {
			log.Println("Error alterting role table", err)
			return
		}
	}
}
