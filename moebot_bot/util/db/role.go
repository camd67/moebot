package db

import (
	"context"
	"database/sql"
	"log"
	"strings"

	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"

	"github.com/camd67/moebot/moebot_bot/util/db/models"
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
							INNER JOIN group_membership ON group_membership.role_id = role.Id
							WHERE group_membership.group_id = $1`
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

func RoleInsertOrUpdateWithoutRoles(role *models.Role) error {
	return RoleInsertOrUpdate(role, nil)
}

func RoleInsertOrUpdate(role *models.Role, groups models.RoleGroupSlice) error {
	r, err := models.Roles(qm.Where("server_id = ? AND role_uid = ?", role.ServerID, role.RoleUID)).One(context.Background(), moeDb)
	if err == sql.ErrNoRows {
		tx, _ := moeDb.Begin()
		if role.Permission == -1 {
			role.Permission = types.PermAll
		}
		r = &models.Role{
			ServerID:   role.ServerID,
			RoleUID:    role.RoleUID,
			Permission: role.Permission,
		}
		err = r.Insert(context.Background(), moeDb, boil.Infer())
		if err != nil {
			log.Println("Error inserting role to db ", err)
			tx.Rollback()
			return err
		}
		if groups != nil {
			err = r.SetRoleGroups(context.Background(), moeDb, false, groups...)
			if err != nil {
				log.Println("Error updating role group relationship to db", err)
				tx.Rollback()
				return err
			}
		}
		tx.Commit()
	} else {
		tx, _ := moeDb.Begin()
		_, err = role.Update(context.Background(), moeDb, boil.Infer())
		if err != nil {
			log.Println("Error updating role", err)
			tx.Rollback()
			return err
		}
		if groups != nil {
			err = role.SetRoleGroups(context.Background(), moeDb, false, groups...)
			if err != nil {
				log.Println("Error updating role group relationship to db", err)
				tx.Rollback()
				return err
			}
		}
		tx.Commit()
	}
	return nil
}

func RoleQueryServer(s *models.Server) (roles models.RoleSlice, err error) {
	roles, err = models.Roles(qm.Where("server_id = ?", s.ID)).All(context.Background(), moeDb)
	if err != nil {
		log.Println("Error querying for role in server role query", err)
	}
	return
}

func RoleQueryGroup(groupId int) (roles models.RoleSlice, err error) {
	roles, err = models.Roles(
		qm.InnerJoin("group_membership gm ON gm.role_id = role.Id"),
		qm.Where("gm.role_group_id = $1", groupId),
	).All(context.Background(), moeDb)
	if err != nil {
		log.Println("Error querying for role in role group query", err)
	}
	return
}

func RoleQueryTrigger(trigger string, serverId int) (r *models.Role, err error) {
	r, err = models.Roles(
		qm.Where("qm.Where(server_id = ? AND UPPER(trigger) = UPPER(?)", serverId, trigger),
	).One(context.Background(), moeDb)
	if err != nil {
		log.Println("Error querying for role by trigger", err)
	}
	return
}

func RoleQueryRoleUid(roleUid string, serverId int) (r *models.Role, err error) {
	r, err = models.Roles(qm.Where("role_uid = ? AND server_uid = ?", roleUid, serverId)).One(context.Background(), moeDb)
	if err == sql.ErrNoRows {
		r = &models.Role{}
	}
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
