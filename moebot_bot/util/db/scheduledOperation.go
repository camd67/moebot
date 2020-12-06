package db

import (
	"context"
	"log"
	"time"

	"github.com/volatiletech/sqlboiler/boil"

	"github.com/volatiletech/sqlboiler/queries"
	"github.com/volatiletech/sqlboiler/queries/qm"

	"github.com/camd67/moebot/moebot_bot/util/db/models"

	"github.com/camd67/moebot/moebot_bot/util/db/types"
)

const (
	SchedulerChannelRotation types.SchedulerType = 1
)

const (
	scheduledOperationTable = `CREATE TABLE IF NOT EXISTS scheduled_operation(
		id SERIAL NOT NULL PRIMARY KEY,
		server_id INTEGER NOT NULL REFERENCES server(id) ON DELETE CASCADE,
		type INTEGER NOT NULL,
		planned_execution_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		execution_interval INTERVAL NOT NULL
	)`

	scheduledOperationQueryNow = `SELECT id, server_id, planned_execution_time FROM scheduled_operation WHERE planned_execution_time < CURRENT_TIMESTAMP`

	scheduledOperationQueryServer = `SELECT id, server_id, planned_execution_time FROM scheduled_operation WHERE server_id = $1`

	scheduledOperationUpdate = `UPDATE scheduled_operation SET planned_execution_time = CURRENT_TIMESTAMP + execution_interval WHERE id = $1 RETURNING planned_execution_time`

	scheduledOperationDelete = `DELETE FROM scheduled_operation WHERE id = $1 AND server_id = $2`

	scheduledOperationInsert = `INSERT INTO scheduled_operation (server_id, type, execution_interval) VALUES ($1, $2, $3) RETURNING id`
)

func scheduledOperationCreateTable() {
	moeDb.Exec(scheduledOperationTable)
}

func ScheduledOperationQueryNow() (models.ScheduledOperationSlice, error) {
	result, err := models.ScheduledOperations(qm.Where("planned_execution_time < CURRENT_TIMESTAMP")).All(context.Background(), moeDb)
	if err != nil {
		log.Println("Error querying for current scheduled operations", err)
		return nil, err
	}
	return result, nil
}

func ScheduledOperationQueryServer(serverID int) (models.ScheduledOperationSlice, error) {
	result, err := models.ScheduledOperations(qm.Where("server_id = ?", serverID)).All(context.Background(), moeDb)
	if err != nil {
		log.Println("Error querying for server scheduled operations", err)
		return nil, err
	}
	return result, nil
}

func ScheduledOperationUpdateTime(operationID int64) (time.Time, error) {
	var sched models.ScheduledOperation
	err := queries.Raw("UPDATE scheduled_operation SET planned_execution_time = CURRENT_TIMESTAMP + execution_interval WHERE id = $1 RETURNING *",
		operationID).Bind(context.Background(), moeDb, &sched)
	if err != nil {
		log.Println("Error updating scheduled operations", err)
		return sched.PlannedExecutionTime, err
	}
	return sched.PlannedExecutionTime, nil
}

func ScheduledOperationDelete(operationID int64, serverID int) (bool, error) {
	rowsAffected, err := models.ScheduledOperations(qm.Where("id = ? AND server_id = ?", operationID, serverID)).DeleteAll(context.Background(), moeDb)
	if err != nil {
		log.Println("Error deleting scheduled operations", err)
		return false, err
	}
	return rowsAffected > 0, err
}

func scheduledOperationInsertNew(serverID int, operationType types.SchedulerType, interval string) (*models.ScheduledOperation, error) {
	operation := &models.ScheduledOperation{
		ServerID:          serverID,
		SchedulerType:     operationType,
		ExecutionInterval: interval,
	}
	err := operation.Insert(context.Background(), moeDb, boil.Infer())
	if err != nil {
		log.Println("Error creating scheduled operation", err)
		return nil, err
	}
	nextExecution, err := ScheduledOperationUpdateTime(int64(operation.ID))
	if err != nil {
		return nil, err
	}
	operation.PlannedExecutionTime = nextExecution
	return operation, nil
}
