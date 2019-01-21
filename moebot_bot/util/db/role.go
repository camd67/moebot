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
	RoleMaxTriggerLength       = 100
	RoleMaxTriggerLengthString = "100"
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
		err = role.Insert(context.Background(), moeDb, boil.Infer())
		if err != nil {
			log.Println("Error inserting role to db ", err)
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
	} else {
		r.ServerID = role.ServerID
		r.RoleUID = role.RoleUID
		r.Permission = role.Permission
		r.Trigger = role.Trigger
		r.ConfirmationMessage = role.ConfirmationMessage
		r.ConfirmationSecurityAnswer = role.ConfirmationSecurityAnswer
		tx, _ := moeDb.Begin()
		_, err = r.Update(context.Background(), moeDb, boil.Infer())
		if err != nil {
			log.Println("Error updating role", err)
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
	}
	return nil
}

func RoleQueryServer(s *models.Server) (roles models.RoleSlice, err error) {
	roles, err = models.Roles(
		qm.Where("server_id = ?", s.ID),
		qm.Load("RoleGroups"),
	).All(context.Background(), moeDb)
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
		qm.Where("server_id = ? AND UPPER(trigger) = UPPER(?)", serverId, trigger),
		qm.Load("RoleGroups"),
	).One(context.Background(), moeDb)
	if err != nil {
		log.Println("Error querying for role by trigger", err)
	}
	return
}

func RoleQueryRoleUid(roleUid string, serverId int) (r *models.Role, err error) {
	r, err = models.Roles(
		qm.Where("role_uid = ? AND server_id = ?", roleUid, serverId),
		qm.Load("RoleGroups"),
	).One(context.Background(), moeDb)
	if err == sql.ErrNoRows {
		r = &models.Role{}
	}
	return
}

func RoleQueryPermission(roleUids []string) (p []types.Permission) {
	convertedUids := make([]interface{}, len(roleUids))
	for index, num := range roleUids {
		convertedUids[index] = num
	}

	roles, err := models.Roles(qm.WhereIn("permission in ?", convertedUids...)).All(context.Background(), moeDb)
	if err != nil {
		log.Println("Error querying for user permissions", err)
		return
	}

	for _, role := range roles {
		p = append(p, role.Permission)
	}
	return
}

func RoleDelete(roleUid string, guildUid string) error {
	s, _ := ServerQueryByGuildUid(guildUid)
	_, err := models.Roles(qm.Where("role_uid = ? AND server_id = ?", roleUid, s.ID)).DeleteAll(context.Background(), moeDb)
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
