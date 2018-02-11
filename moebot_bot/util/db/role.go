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
	// Used to disable something, no one can have this permission level
	PermNone Permission = 100
	// Master level permission, can't ever be ignored or disabled
	PermMaster Permission = 101
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
	RoleCustom
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

	roleQueryServerRole  = `SELECT Id, ServerId, RoleUid, Permission, RoleType FROM role WHERE ServerId = $1 AND RoleUid = $2`
	roleQuery            = `SELECT Id, ServerId, RoleUid, Permission, RoleType FROM role WHERE Id = $1`
	roleQueryPermissions = `SELECT Permission FROM role WHERE RoleUid = ANY ($1::varchar[])`
	roleUpdate           = `UPDATE role SET Permission = $1, RoleType = $2 WHERE Id = $3`
	roleInsert           = `INSERT INTO role(ServerId, RoleUid, Permission, RoleType) VALUES($1, $2, $3, $4) RETURNING id`
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
			_, err = moeDb.Exec(roleInsert, role.ServerId, strings.TrimSpace(role.RoleUid), role.Permission, role.RoleType)
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

func RoleQueryOrInsert(role Role) (r Role, err error) {
	row := moeDb.QueryRow(roleQueryServerRole, role.ServerId, role.RoleUid)
	if err = row.Scan(&r.Id, &r.ServerId, &r.RoleUid, &r.Permission, &r.RoleType); err != nil {
		if err == sql.ErrNoRows {
			// no row, so insert it add in default values
			if role.Permission == -1 {
				role.Permission = PermNone
			}
			if role.RoleType == -1 {
				role.RoleType = RoleNone
			}
			var insertId int
			err = moeDb.QueryRow(roleInsert, role.ServerId, strings.TrimSpace(role.RoleUid), role.Permission, role.RoleType).Scan(&insertId)
			if err != nil {
				log.Println("Error inserting role to db")
				return
			}
			row := moeDb.QueryRow(roleQuery, insertId)
			if err = row.Scan(&r.Id, &r.ServerId, &r.RoleUid, &r.Permission, &r.RoleType); err != nil {
				log.Println("Failed to read the newly inserted Role row. This should pretty much never happen...", err)
				return Role{}, err
			}
		}
	}
	// got a row, return it
	return
}

func RoleQueryPermission(roleUid []string) (p []Permission) {
	idCollection := "{" + strings.Join(roleUid, ",") + "}"
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

func GetPermissionFromString(s string) Permission {
	toCheck := strings.ToUpper(s)
	if toCheck == "ALL" {
		return PermAll
	} else if toCheck == "MOD" {
		return PermMod
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
	case PermMod:
		return "Mod"
	case PermAll:
		return "All"
	case PermNone:
		return "None"
	case PermMaster:
		return "Master"
	default:
		return "Unknown"
	}
}
