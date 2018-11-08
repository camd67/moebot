package db

import (
	"database/sql"
	"log"
)

const (
	groupMembershipTable = `CREATE TABLE IF NOT EXISTS group_membership(
		role_id INTEGER NOT NULL REFERENCES role(Id) ON DELETE CASCADE,
		group_id INTEGER NOT NULL REFERENCES role_group(Id) ON DELETE CASCADE,
		CONSTRAINT group_membership_pkey PRIMARY KEY (role_id, group_id)
	)`

	groupMembershipQueryByRole  = `SELECT group_id FROM group_membership WHERE role_id = $1`
	groupMembershipQueryByGroup = `SELECT group_id FROM group_membership WHERE group_id = $1`
	groupMembershipInsert       = `INSERT INTO group_membership(role_id, group_id) VALUES($1, $2)`
	groupMembershipDelete       = `DELETE FROM group_membership WHERE role_id = $1 AND group_id = $2`
)

func groupMembershipQueryByRoleID(roleID int) ([]int, error) {
	rows, err := moeDb.Query(groupMembershipQueryByRole, roleID)
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

func groupMembershipQueryByGroupID(groupID int) ([]int, error) {
	rows, err := moeDb.Query(groupMembershipQueryByGroup, groupID)
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

func groupMembershipAdd(roleID int, groupID int) error {
	_, err := moeDb.Exec(groupMembershipInsert, roleID, groupID)
	return err
}

func groupMembershipRemove(roleID int, groupID int) error {
	_, err := moeDb.Exec(groupMembershipDelete, roleID, groupID)
	return err
}

func groupMembershipCreateTable() {
	_, err := moeDb.Exec(groupMembershipTable)
	if err != nil {
		log.Println("Error creating role group relation table", err)
		return
	}
}
