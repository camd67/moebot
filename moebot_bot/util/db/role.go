package db

import (
	"database/sql"
	"log"
	"strconv"
	"strings"

	"github.com/camd67/moebot/moebot_bot/util/db/types"
)

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

	roleQueryServerRole = `SELECT Id, ServerId, RoleUid, Permission, ConfirmationMessage, ConfirmationSecurityAnswer, Trigger FROM role WHERE RoleUid = $1 AND ServerId = $2`
	roleQueryServer     = `SELECT Id, ServerId, RoleUid, Permission, ConfirmationMessage, ConfirmationSecurityAnswer, Trigger FROM role WHERE ServerId = $1`
	roleQuery           = `SELECT Id, ServerId, RoleUid, Permission, ConfirmationMessage, ConfirmationSecurityAnswer, Trigger FROM role WHERE Id = $1`
	roleQueryTrigger    = `SELECT Id, ServerId, RoleUid, Permission, ConfirmationMessage, ConfirmationSecurityAnswer, Trigger FROM role WHERE UPPER(Trigger) = UPPER($1) AND ServerId = $2`
	roleQueryGroup      = `SELECT Id, ServerId, RoleUid, Permission, ConfirmationMessage, ConfirmationSecurityAnswer, Trigger FROM role 
							INNER JOIN role_group_role ON role_group_role.role_id = role.Id
							WHERE role_group_role.group_id = $1`
	roleQueryPermissions = `SELECT Permission FROM role WHERE RoleUid = ANY ($1::varchar[])`

	roleUpdate = `UPDATE role SET Permission = $2, ConfirmationMessage = $3, ConfirmationSecurityAnswer = $4, Trigger = $5 WHERE Id = $1`

	roleInsert = `INSERT INTO role(ServerId, RoleUid, Permission, ConfirmationMessage, ConfirmationSecurityAnswer, Trigger) VALUES($1, $2, $3, $4, $5, $6) RETURNING id`

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

func RoleInsertOrUpdate(role types.Role) error {
	row := moeDb.QueryRow(roleQueryServerRole, role.RoleUid, role.ServerId)
	var r types.Role
	if err := row.Scan(&r.Id, &r.ServerId, &r.RoleUid, &r.Permission, &r.ConfirmationMessage, &r.ConfirmationSecurityAnswer, &r.Trigger); err != nil {
		if err == sql.ErrNoRows {
			// no row, so insert it add in default values
			if role.Permission == -1 {
				role.Permission = types.PermAll
			}
			tx, _ := moeDb.Begin()
			var insertID int
			err = moeDb.QueryRow(roleInsert, role.ServerId, strings.TrimSpace(role.RoleUid), role.Permission, role.ConfirmationMessage,
				role.ConfirmationSecurityAnswer, role.Trigger).Scan(&insertID)
			if err != nil {
				log.Println("Error inserting role to db", err)
				tx.Rollback()
				return err
			}
			for _, groupID := range role.Groups {
				err = roleGroupRelationAdd(insertID, groupID)
				if err != nil {
					log.Println("Error inserting role group relationship to db", err)
					tx.Rollback()
					return err
				}
			}
			tx.Commit()
		} else {
			log.Println("Error scanning for role", err)
			return err
		}
	} else {
		// got a row, update it
		r.Groups, err = roleGroupRelationQueryRole(r.Id)
		if err != nil && err != sql.ErrNoRows {
			log.Println("Error scanning for role group relationships", err)
			return err
		}
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
		tx, _ := moeDb.Begin()
		_, err = moeDb.Exec(roleUpdate, r.Id, r.Permission, r.ConfirmationMessage, r.ConfirmationSecurityAnswer, r.Trigger)
		if err != nil {
			log.Println("Error updating role to db: Id "+strconv.Itoa(r.Id), err)
			tx.Rollback()
			return err
		}
		groupsToAdd := subtract(role.Groups, r.Groups)
		for _, groupID := range groupsToAdd {
			err = roleGroupRelationAdd(r.Id, groupID)
			if err != nil {
				log.Println("Error inserting role group relationship to db", err)
				tx.Rollback()
				return err
			}
		}
		groupsToRemove := subtract(r.Groups, role.Groups)
		for _, groupID := range groupsToRemove {
			err = roleGroupRelationRemove(r.Id, groupID)
			if err != nil {
				log.Println("Error removing role group relationship to db", err)
				tx.Rollback()
				return err
			}
		}
		tx.Commit()
	}
	return nil
}

func RoleQueryOrInsert(role types.Role) (r types.Role, err error) {
	row := moeDb.QueryRow(roleQueryServerRole, role.ServerId, role.RoleUid)
	if err = row.Scan(&r.Id, &r.ServerId, &r.RoleUid, &r.Permission, &r.ConfirmationMessage, &r.ConfirmationSecurityAnswer, &r.Trigger); err != nil {
		if err == sql.ErrNoRows {
			// no row, so insert it add in default values
			if role.Permission == -1 {
				role.Permission = types.PermAll
			}
			tx, _ := moeDb.Begin()
			err = moeDb.QueryRow(roleInsert, role.ServerId, strings.TrimSpace(role.RoleUid), role.Permission, role.ConfirmationMessage,
				role.ConfirmationSecurityAnswer, role.Trigger).Scan(&role.Id)
			if err != nil {
				log.Println("Error inserting role to db")
				tx.Rollback()
				return
			}
			for _, groupID := range role.Groups {
				err = roleGroupRelationAdd(role.Id, groupID)
				if err != nil {
					log.Println("Error inserting role group relationship to db", err)
					tx.Rollback()
					return
				}
			}
			tx.Commit()
			r = role
		}
	} else {
		r.Groups, err = roleGroupRelationQueryRole(r.Id)
	}
	// got a row, return it
	return
}

func RoleQueryServer(s types.Server) (roles []types.Role, err error) {
	rows, err := moeDb.Query(roleQueryServer, s.Id)
	if err != nil {
		log.Println("Error querying for role", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var r types.Role
		if err = rows.Scan(&r.Id, &r.ServerId, &r.RoleUid, &r.Permission, &r.ConfirmationMessage, &r.ConfirmationSecurityAnswer,
			&r.Trigger); err != nil {

			log.Println("Error scanning from role table:", err)
			return
		}
		if r.Groups, err = roleGroupRelationQueryRole(r.Id); err != nil {
			log.Println("Error scanning from role group relation table:", err)
			return
		}
		roles = append(roles, r)
	}
	return
}

func RoleQueryGroup(groupId int) (roles []types.Role, err error) {
	rows, err := moeDb.Query(roleQueryGroup, groupId)
	if err != nil {
		log.Println("Error querying for role", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var r types.Role
		if err = rows.Scan(&r.Id, &r.ServerId, &r.RoleUid, &r.Permission, &r.ConfirmationMessage, &r.ConfirmationSecurityAnswer,
			&r.Trigger); err != nil {

			log.Println("Error scanning from role table:", err)
			return
		}
		if r.Groups, err = roleGroupRelationQueryRole(r.Id); err != nil {
			log.Println("Error scanning from role group relation table:", err)
			return
		}
		roles = append(roles, r)
	}
	return
}

func RoleQueryTrigger(trigger string, serverId int) (r types.Role, err error) {
	row := moeDb.QueryRow(roleQueryTrigger, trigger, serverId)
	err = row.Scan(&r.Id, &r.ServerId, &r.RoleUid, &r.Permission, &r.ConfirmationMessage, &r.ConfirmationSecurityAnswer, &r.Trigger)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error querying for role by trigger", err)
	}
	if r.Groups, err = roleGroupRelationQueryRole(r.Id); err != nil {
		log.Println("Error scanning from role group relation table:", err)
		return
	}
	// return whatever we get, error or row
	return
}

func RoleQueryRoleUid(roleUid string, serverId int) (r types.Role, err error) {
	row := moeDb.QueryRow(roleQueryServerRole, roleUid, serverId)
	err = row.Scan(&r.Id, &r.ServerId, &r.RoleUid, &r.Permission, &r.ConfirmationMessage, &r.ConfirmationSecurityAnswer, &r.Trigger)
	if err == nil {
		if r.Groups, err = roleGroupRelationQueryRole(r.Id); err != nil {
			log.Println("Error scanning from role group relation table:", err)
		}
	} else {
		if err != sql.ErrNoRows {
			log.Println("Error querying for role by UID and serverID", err)
		}
	}
	// return whatever we get, error or row
	return
}

func RoleQueryPermission(roleUids []string) (p []types.Permission) {
	idCollection := "{" + strings.Join(roleUids, ",") + "}"
	r, err := moeDb.Query(roleQueryPermissions, idCollection)
	if err != nil {
		log.Println("Error querying for user permissions", err)
		return
	}
	for r.Next() {
		var newPerm types.Permission
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

/*
Gets a permission value from a string. This should be used when accepting user input.
*/
func GetPermissionFromString(s string) types.Permission {
	toCheck := strings.ToUpper(s)
	if toCheck == "ALL" {
		return types.PermAll
	} else if toCheck == "MOD" {
		return types.PermMod
	} else if toCheck == "GUILD OWNER" || toCheck == "GO" {
		return types.PermGuildOwner
	} else if toCheck == "NONE" {
		return types.PermNone
	} else if toCheck == "MASTER" {
		return types.PermMaster
	} else {
		return -1
	}
}

/*
Gets a string from a permission level for use when informing users of what permission they can enter
*/
func SprintPermission(p types.Permission) string {
	switch p {
	case types.PermAll:
		return "All"
	case types.PermMod:
		return "Mod"
	case types.PermGuildOwner:
		return "Guild Owner"
	case types.PermNone:
		return "None"
	case types.PermMaster:
		return "Master"
	default:
		return "Unknown"
	}
}

/*
Gets a string from a permission, which is the user-facing string NOT the assignable string.
For example: "Your permission level is: GetPermissionString(PermAll)"
*/
func GetPermissionString(p types.Permission) string {
	switch p {
	case types.PermAll:
		return "Normal User"
	case types.PermMod:
		return "Mod"
	case types.PermGuildOwner:
		return "Guild Owner"
	case types.PermNone:
		return "How did you get this role...?"
	case types.PermMaster:
		return "Master"
	default:
		return "Unknown"
	}
}

/*
Currently only a subset of roles are assignable by the bot
*/
func IsAssignablePermissionLevel(p types.Permission) bool {
	return p == types.PermMod || p == types.PermAll
}

/*
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

func roleDatabaseUpdate() {
	const veteranUpdate = "INSERT INTO role(ServerId, RoleUid, Permission, Trigger) SELECT server.Id, server.VeteranRole, 1, 'veteran' FROM server WHERE server.VeteranRole IS NOT NULL AND server.VeteranRole <> '' ON CONFLICT DO NOTHING"
	_, err := moeDb.Exec(veteranUpdate)
	if err != nil {
		log.Println("Error updating role table", err)
		return
	}
	const oldRoleSelectGroups = "SELECT Id, GroupId FROM role WHERE GroupId <> 0"
	const insertRoleRelation = "INSERT INTO role_group_role(role_id, group_id) VALUES($1,$2) ON CONFLICT DO NOTHING"
	const oldRoleRemoveRelation = "UPDATE role SET GroupId = 0 WHERE Id = $1"
	rows, err := moeDb.Query(oldRoleSelectGroups)
	if err != nil {
		log.Println("Error updating role groups", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var roleID int
		var groupID int
		if err = rows.Scan(&roleID, &groupID); err != nil {
			log.Println("Error scanning from role table:", err)
			return
		}
		_, err = moeDb.Exec(insertRoleRelation, roleID, groupID)
		if err == nil {
			moeDb.Exec(oldRoleRemoveRelation, roleID)
		}
	}
}
