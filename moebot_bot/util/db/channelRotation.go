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
		channel_uids VARCHAR(MAX) NOT NULL
	)`

	channelRotationQuery = `SELECT channel_rotation.operation_id, channel_rotation.current_channel_uid, channel_rotation.channel_uids, scheduled_operation.server_id 
							FROM channel_rotation 
							INNER JOIN scheduled_operation ON scheduled_operation.id = channel_rotation.operation_id
							WHERE operation_id = $1`

	channelRotationUpdate = `UPDATE channel_rotation SET current_channel_uid = $2 WHERE operation_id = $1`
)

func ChannelRotationQuery(operationID int) (*ChannelRotation, error) {
	cr := &ChannelRotation{}
	channelList := ""
	row := moeDb.QueryRow(channelRotationQuery, operationID)
	if e := row.Scan(&cr.ID, &cr.CurrentChannelUID, &channelList, &cr.ServerID); e != nil {
		return nil, e
	}
	cr.ChannelUIDList = strings.Split(channelList, " ")
	return cr, nil
}

func ChannelRotationUpdate(operationID int, currentChannelUID string) error {
	_, err := moeDb.Exec(channelRotationUpdate, operationID, currentChannelUID)
	if err != nil {
		log.Println("Error updating channel rotation", err)
		return err
	}
	return nil
}
