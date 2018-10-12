package db

import (
	"log"
	"strings"
)

type ChannelRotation struct {
	ChannelUIDList    []string
	CurrentChannelUID string
	ScheduledOperation
}

const (
	channelRotationTable = `CREATE TABLE IF NOT EXISTS channel_rotation(
		operation_id INTEGER NOT NULL PRIMARY KEY REFERENCES scheduled_operation(id) ON DELETE CASCADE,
		current_channel_uid VARCHAR(20) NOT NULL,
		channel_uids VARCHAR(1000) NOT NULL
	)`

	channelRotationQuery = `SELECT channel_rotation.operation_id, channel_rotation.current_channel_uid, channel_rotation.channel_uids, scheduled_operation.server_id 
							FROM channel_rotation 
							INNER JOIN scheduled_operation ON scheduled_operation.id = channel_rotation.operation_id
							WHERE operation_id = $1`

	channelRotationUpdate = `UPDATE channel_rotation SET current_channel_uid = $2 WHERE operation_id = $1`

	channelRotationInsert = `INSERT INTO channel_rotation (operation_id, current_channel_uid, channel_uids) VALUES($1, $2, $3)`
)

func channelRotationCreateTable() {
	moeDb.Exec(channelRotationTable)
}

func ChannelRotationQuery(operationID int64) (*ChannelRotation, error) {
	cr := &ChannelRotation{}
	channelList := ""
	row := moeDb.QueryRow(channelRotationQuery, operationID)
	if e := row.Scan(&cr.ID, &cr.CurrentChannelUID, &channelList, &cr.ServerID); e != nil {
		return nil, e
	}
	cr.ChannelUIDList = strings.Split(channelList, " ")
	return cr, nil
}

func ChannelRotationUpdate(operationID int64, currentChannelUID string) error {
	_, err := moeDb.Exec(channelRotationUpdate, operationID, currentChannelUID)
	if err != nil {
		log.Println("Error updating channel rotation", err)
		return err
	}
	return nil
}

func ChannelRotationAdd(serverID int, currentChannelUID string, channels []string, interval string) error {
	operation, err := scheduledOperationInsertNew(serverID, SchedulerChannelRotation, interval)
	if err != nil {
		return err
	}
	_, err = moeDb.Exec(channelRotationInsert, operation.ID, currentChannelUID, strings.Join(channels, " "))
	return err
}

func (c *ChannelRotation) NextChannelUID() string {
	if len(c.ChannelUIDList) == 0 {
		return ""
	}
	if len(c.ChannelUIDList) == 1 {
		//this way, we handle single channels "rotations", which is a channel being visible and hidden on a set amount of time
		if c.CurrentChannelUID == "" {
			return c.ChannelUIDList[0]
		}
		return ""
	}
	var nextIndex int
	for i := 0; i < len(c.ChannelUIDList); i++ {
		if c.CurrentChannelUID == c.ChannelUIDList[i] {
			nextIndex = i + 1
			break
		}
	}
	if nextIndex >= len(c.ChannelUIDList) {
		nextIndex = 0
	}
	return c.ChannelUIDList[nextIndex]
}
