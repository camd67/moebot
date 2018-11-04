package db

import (
	"database/sql"
	"log"
)

const (
	roleGroupRoleTable = `CREATE TABLE IF NOT EXISTS role_group_role(
		role_id INTEGER NOT NULL REFERENCES role(Id) ON DELETE CASCADE,
		group_id INTEGER NOT NULL REFERENCES role_group(Id) ON DELETE CASCADE,
		CONSTRAINT role_group_role_pkey PRIMARY KEY (role_id, group_id)
	)`

	roleGroupRoleQueryRole  = `SELECT group_id FROM role_group_role WHERE role_id = $1`
	roleGroupRoleQueryGroup = `SELECT group_id FROM role_group_role WHERE group_id = $1`
	roleGroupRoleAdd        = `INSERT INTO role_group_role(role_id, group_id) VALUES($1, $2)`
	roleGroupRoleRemove     = `DELETE FROM role_group_role WHERE role_id = $1 AND group_id = $2`
)

func roleGroupRelationQueryRole(roleID int) ([]int, error) {
	rows, err := moeDb.Query(roleGroupRoleQueryRole, roleID)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error querying for role groups", err)
			return nil, err
		}
		return []int{}, nil
	}
	result := []int{}
	for rows.Next() {
		var groupID int
		err = rows.Scan(&groupID)
		if err != nil {
			log.Println("Error querying for role groups", err)
			return nil, err
		}
		result = append(result, groupID)
	}
	return result, nil
}

func roleGroupRelationQueryGroup(groupID int) ([]int, error) {
	rows, err := moeDb.Query(roleGroupRoleQueryGroup, groupID)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Println("Error querying for group roles", err)
			return nil, err
		}
		return []int{}, nil
	}
	result := []int{}
	for rows.Next() {
		var roleID int
		err = rows.Scan(&roleID)
		if err != nil {
			log.Println("Error querying for group roles", err)
			return nil, err
		}
		result = append(result, roleID)
	}
	return result, nil
}

func roleGroupRelationAdd(roleID int, groupID int) error {
	_, err := moeDb.Exec(roleGroupRoleAdd, roleID, groupID)
	return err
}

func roleGroupRelationRemove(roleID int, groupID int) error {
	_, err := moeDb.Exec(roleGroupRoleRemove, roleID, groupID)
	return err
}

func subtract(slice1 []int, slice2 []int) []int {
	var result []int
	for _, a := range slice1 {
		if !contains(slice2, a) {
			result = append(result, a)
		}
	}
	return result
}

func contains(slice []int, value int) bool {
	for _, a := range slice {
		if a == value {
			return true
		}
	}
	return false
}

func roleGroupRoleCreateTable() {
	_, err := moeDb.Exec(roleGroupRoleTable)
	if err != nil {
		log.Println("Error creating role group relation table", err)
		return
	}
}
