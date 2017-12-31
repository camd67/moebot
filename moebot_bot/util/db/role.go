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
	_ Permission = iota
	PermAll
	PermMod
	PermNone
)

// RoleType enum
type RoleType int

const (
	_ RoleType = iota
	// the starter role you get when joining a server (if enabled)
	RoleStarter
	// the default role you get AFTER agreeing to the rules
	RoleDefault
	RoleRank
	RoleTeam
	RoleNone
)

type Role struct {
	Id         int
	ServerId   int
	RoleUid    string
	Permission Permission
	RoleType   RoleType
}

const (
	roleTable = `CREATE TABLE IF NOT EXISTS role(
		Id SERIAL NOT NULL PRIMARY KEY,
		ServerId INTEGER REFERENCES server(Id) ON DELETE CASCADE,
		RoleUid VARCHAR(20) NOT NULL UNIQUE,
		Permission SMALLINT NOT NULL,
		RoleType SMALLINT NOT NULL
	)`

	roleQueryServerRole = `SELECT Id, ServerId, RoleUid, Permission, RoleType FROM role WHERE ServerId = $1 AND RoleUid = $2`
	roleUpdate          = `UPDATE role SET Permission = $1, RoleType = $2 WHERE Id = $3`
	roleInsert          = `INSERT INTO role(ServerId, RoleUid, Permission, RoleType) VALUES($1, $2, $3, $4)`
)

func RoleInsertOrUpdate(role Role) error {
	row := moeDb.QueryRow(roleQueryServerRole, role.ServerId, role.RoleUid)
	var r Role
	if err := row.Scan(&r.Id, &r.ServerId, &r.RoleUid, &r.Permission, &r.RoleType); err != nil {
		if err == sql.ErrNoRows {
			// no row, so insert it add in default values
			if role.Permission == -1 {
				role.Permission = PermNone
			}
			if role.RoleType == -1 {
				role.RoleType = RoleNone
			}
			_, err = moeDb.Exec(roleInsert, role.ServerId, role.RoleUid, role.Permission, role.RoleType)
			if err != nil {
				log.Println("Error inserting role to db")
				return err
			}
		}
	} else {
		// got a row, update it
		if role.Permission != -1 {
			r.Permission = role.Permission
		}
		if role.RoleType != -1 {
			r.RoleType = role.RoleType
		}
		_, err = moeDb.Exec(roleUpdate, r.Permission, r.RoleType, r.Id)
		if err != nil {
			log.Println("Error updating role to db: Id - " + strconv.Itoa(r.Id))
			return err
		}
	}
	return nil
}

func GetPermissionFromString(s string) Permission {
	toCheck := strings.ToUpper(s)
	if toCheck == "ALL" {
		return PermAll
	} else if toCheck == "MOD" {
		return PermMod
	} else if toCheck == "NONE" {
		return PermNone
	} else {
		return -1
	}
}

func SprintPermission(p Permission) string {
	switch p {
	case PermMod:
		return "Mod"
	case PermAll:
		return "All"
	case PermNone:
		return "None"
	default:
		return "Unknown"
	}
}
