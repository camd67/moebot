package db

import (
	"log"
	"time"
)

type SchedulerType int

const (
	SchedulerChannelRotation SchedulerType = iota
)

type ScheduledOperation struct {
	ID                   int64
	ServerID             int
	Type                 SchedulerType
	PlannedExecutionTime time.Time
}

const (
	scheduledOperationTable = `CREATE TABLE IF NOT EXISTS scheduled_operation(
		id SERIAL NOT NULL PRIMARY KEY,
		server_id INTEGER NOT NULL REFERENCES server(id) ON DELETE CASCADE,
		type INTEGER NOT NULL,
		planned_execution_time TIMESTAMP NOT NULL,
		execution_interval INTERVAL NOT NULL
	)`

	scheduledOperationQueryNow = `SELECT id, server_id, planned_execution_time FROM channel_rotation WHERE planned_execution_time < CURRENT_TIMESTAMP`

	scheduledOperationUpdate = `UPDATE scheduled_operation SET planned_execution_time = planned_execution_time + execution_interval WHERE id = $1`

	scheduledOperationDelete = `DELETE FROM scheduled_operation WHERE id = $1`
)

func ScheduledOperationQueryNow() ([]*ScheduledOperation, error) {
	rows, err := moeDb.Query(scheduledOperationQueryNow)
	if err != nil {
		log.Println("Error querying for scheduled operations", err)
		return nil, err
	}
	result := []*ScheduledOperation{}
	for rows.Next() {
		operation := new(ScheduledOperation)
		err = rows.Scan(&operation.ID, &operation.ServerID, &operation.PlannedExecutionTime)
		if err != nil {
			log.Println("Error querying for scheduled operations", err)
			return nil, err
		}
		result = append(result, operation)
	}
	return result, nil
}

func ScheduledOperationUpdateTime(operationID int) error {
	_, err := moeDb.Exec(scheduledOperationUpdate, operationID)
	if err != nil {
		log.Println("Error updating scheduled operations", err)
		return err
	}
	return nil
}

func ScheduledOperationDelete(operationID int) error {
	_, err := moeDb.Exec(scheduledOperationDelete, operationID)
	if err != nil {
		log.Println("Error deleting scheduled operations", err)
		return err
	}
	return nil
}
