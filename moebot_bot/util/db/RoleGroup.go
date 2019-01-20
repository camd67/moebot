package db

import (
	"context"
	"database/sql"
	"log"
	"strconv"
	"strings"

	"github.com/volatiletech/sqlboiler/boil"

	"github.com/camd67/moebot/moebot_bot/util/db/models"
	"github.com/camd67/moebot/moebot_bot/util/db/types"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

const (
	RoleGroupMaxNameLength       = 500
	RoleGroupMaxNameLengthString = "500"

	UncategorizedGroup = "Uncategorized"
)

func RoleGroupInsertOrUpdate(rg *models.RoleGroup, s *models.Server) (id int, err error) {
	rg.ServerID = s.ID
	if _, err := models.FindRoleGroup(context.Background(), moeDb, rg.ID); err != nil {
		if err == sql.ErrNoRows {
			if rg.GroupType <= 0 {
				rg.GroupType = types.GroupTypeAny
			}
			err = rg.Insert(context.Background(), moeDb, boil.Infer())
			if err != nil {
				log.Println("Error inserting roleGroup to db")
				return -1, err
			}
		} else {
			// got some other kind of error
			log.Println("Error scanning roleGroup row from database", err)
			return -1, err
		}
	} else {
		_, err = rg.Update(context.Background(), moeDb, boil.Infer())
		if err != nil {
			log.Println("Error updating roleGroup to db: Id - " + strconv.Itoa(rg.ID))
			return -1, err
		}
	}
	return rg.ID, nil
}

func RoleGroupQueryServer(s *models.Server) (roleGroups models.RoleGroupSlice, err error) {
	roleGroups, err = models.RoleGroups(qm.Where("server_id = ?", s.ID)).All(context.Background(), moeDb)
	if err != nil {
		log.Println("Error querying for roleGroup", err)
	}
	return
}

func RoleGroupQueryName(name string, serverId int) (rg *models.RoleGroup, err error) {
	rg, err = models.RoleGroups(qm.Where("name = ? AND server_id = ?", name, serverId)).One(context.Background(), moeDb)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error querying for role group by name and serverID", err)
	}
	// return whatever we get, error or row
	return
}

func RoleGroupQueryId(id int) (rg *models.RoleGroup, err error) {
	rg, err = models.FindRoleGroup(context.Background(), moeDb, id)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error querying for role group by id", err)
	}
	return
}

func RoleGroupDelete(id int) error {
	_, err := models.RoleGroups(qm.Where("ID = ?", id)).DeleteAll(context.Background(), moeDb)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error deleting role group: ", id)
	}
	return err
}

func GetGroupTypeFromString(s string) types.GroupType {
	toCheck := strings.ToUpper(s)
	if toCheck == "ANY" {
		return types.GroupTypeAny
	} else if toCheck == "EXCLUSIVE" || toCheck == "EXC" {
		return types.GroupTypeExclusive
	} else if toCheck == "EXCLUSIVE NO REMOVE" || toCheck == "ENR" {
		return types.GroupTypeExclusiveNoRemove
	} else if toCheck == "NO MULTIPLES" || toCheck == "NOM" {
		return types.GroupTypeNoMultiples
	} else {
		return -1
	}
}

func GetStringFromGroupType(groupType types.GroupType) string {
	switch groupType {
	case types.GroupTypeAny:
		return "Any (ANY)"
	case types.GroupTypeExclusive:
		return "Exclusive (EXC)"
	case types.GroupTypeExclusiveNoRemove:
		return "Exclusive No Remove (ENR)"
	case types.GroupTypeNoMultiples:
		return "No Multiples (NOM)"
	default:
		return "Unknown"
	}
}
