package db

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
)

type GroupType int

const (
	//GroupTypeAny Group type where any role can be selected, and multiple can be selected
	GroupTypeAny = 1
	//GroupTypeExclusive Group type where only one role can be selected
	GroupTypeExclusive = 2
	//GroupTypeExclusiveNoRemove Same as the exclusive group, but can't be removed
	GroupTypeExclusiveNoRemove = 3
	//GroupTypeEvent Role is assigned through event points
	GroupTypeEvent = 4

	OptionsForGroupType = "ANY, EXC, ENR"
)

type RoleGroup struct {
	Id       int
	ServerId int
	Name     string
	Type     GroupType
}

const (
	roleGroupTable = `CREATE TABLE IF NOT EXISTS role_group(
			Id SERIAL NOT NULL PRIMARY KEY,
			ServerId INTEGER NOT NULL REFERENCES Server(Id),
			Name TEXT NOT NULL CHECK(char_length(Name) <= 500),
			Type INTEGER NOT NULL
		)`

	RoleGroupMaxNameLength       = 500
	RoleGroupMaxNameLengthString = "500"

	roleGroupQueryById     = `SELECT Id, ServerId, Name, Type FROM role_group WHERE Id = $1`
	roleGroupQueryByName   = `SELECT rg.Id, rg.ServerId, rg.Name, rg.Type FROM role_group AS rg WHERE rg.Name = $1 AND rg.ServerId = $2`
	roleGroupQueryByServer = `SELECT Id, ServerId, Name, Type FROM role_group WHERE ServerId = $1`
	roleGroupInsert        = `INSERT INTO role_group(ServerId, Name, Type) VALUES ($1, $2, $3) RETURNING Id`
	roleGroupUpdate        = `UPDATE role_group SET Name = $2, Type = $3 WHERE Id = $1`
	roleGroupDeleteId      = `DELETE FROM role_group WHERE Id = $1`

	UncategorizedGroup = "Uncategorized"
)

func RoleGroupInsertOrUpdate(rg RoleGroup, s Server) (newId int, err error) {
	row := moeDb.QueryRow(roleGroupQueryById, rg.Id)
	var dbRg RoleGroup
	if err := row.Scan(&dbRg.Id, &dbRg.ServerId, &dbRg.Name, &dbRg.Type); err != nil {
		if err == sql.ErrNoRows {
			// no row, so insert it add in default values
			if rg.Type <= 0 {
				rg.Type = GroupTypeAny
			}
			err := moeDb.QueryRow(roleGroupInsert, s.Id, rg.Name, rg.Type).Scan(&newId)
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
		// got a row, update it
		if rg.Type > 0 {
			dbRg.Type = rg.Type
		}
		if rg.Name != "" {
			dbRg.Name = rg.Name
		}
		_, err = moeDb.Exec(roleGroupUpdate, dbRg.Id, dbRg.Name, dbRg.Type)
		if err != nil {
			log.Println("Error updating roleGroup to db: Id - " + strconv.Itoa(dbRg.Id))
			return -1, err
		}
		newId = dbRg.Id
	}
	return newId, nil
}

/*
Returns a RoleGroup matching the id inside the given RoleGroup. If no match is found, the RoleGroup is added to the database
*/
func RoleGroupQueryOrInsert(rg RoleGroup, s Server) (newRg RoleGroup, err error) {
	row := moeDb.QueryRow(roleGroupQueryById, rg.Id)
	if err = row.Scan(&newRg.Id, &newRg.ServerId, &newRg.Name, &newRg.Type); err != nil {
		if err == sql.ErrNoRows {
			// no row, so insert it add in default values
			if rg.Type <= 0 {
				rg.Type = GroupTypeAny
			}
			var insertId int
			err = moeDb.QueryRow(roleGroupInsert, s.Id, rg.Name, rg.Type).Scan(&insertId)
			if err != nil {
				log.Println("Error inserting role to db")
				return
			}
			// no need to re-query since we inserted a row
			newRg.Id = insertId
		} else {
			log.Println("Error scanning in roleGroup", err)
			return RoleGroup{}, err
		}
	}
	return
}

func RoleGroupQueryServer(s Server) (roleGroups []RoleGroup, err error) {
	rows, err := moeDb.Query(roleGroupQueryByServer, s.Id)
	if err != nil {
		log.Println("Error querying for roleGroup", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var rg RoleGroup
		if err = rows.Scan(&rg.Id, &rg.ServerId, &rg.Name, &rg.Type); err != nil {
			log.Println("Error scanning from roleGroup table:", err)
			return
		}
		roleGroups = append(roleGroups, rg)
	}
	return
}

func RoleGroupQueryName(name string, serverId int) (rg RoleGroup, err error) {
	row := moeDb.QueryRow(roleGroupQueryByName, name, serverId)
	err = row.Scan(&rg.Id, &rg.ServerId, &rg.Name, &rg.Type)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error querying for role group by name and serverID", err)
	}
	// return whatever we get, error or row
	return
}

func RoleGroupQueryId(id int) (rg RoleGroup, err error) {
	row := moeDb.QueryRow(roleGroupQueryById, id)
	err = row.Scan(&rg.Id, &rg.ServerId, &rg.Name, &rg.Type)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error querying for role group by id", err)
	}
	return
}

func RoleGroupDelete(id int) error {
	_, err := moeDb.Exec(roleGroupDeleteId, id)
	if err != nil && err != sql.ErrNoRows {
		log.Println("Error deleting role group: ", id)
	}
	return err
}

func roleGroupCreateTable() {
	_, err := moeDb.Exec(roleGroupTable)
	if err != nil {
		log.Println("Error creating role group table", err)
		return
	}
	//for _, alter := range roleUpdateTable {
	//	_, err = moeDb.Exec(alter)
	//	if err != nil {
	//		log.Println("Error alterting role group table", err)
	//		return
	//	}
	//}
}

func GetGroupTypeFromString(s string) GroupType {
	toCheck := strings.ToUpper(s)
	if toCheck == "ANY" {
		return GroupTypeAny
	} else if toCheck == "EXCLUSIVE" || toCheck == "EXC" {
		return GroupTypeExclusive
	} else if toCheck == "EXCLUSIVE NO REMOVE" || toCheck == "ENR" {
		return GroupTypeExclusiveNoRemove
	} else {
		return -1
	}
}

func GetStringFromGroupType(groupType GroupType) string {
	switch groupType {
	case GroupTypeAny:
		return "Any (ANY)"
	case GroupTypeExclusive:
		return "Exclusive (EXC)"
	case GroupTypeExclusiveNoRemove:
		return "Exclusive No Remove (ENR)"
	default:
		return "Unknown"
	}
}
