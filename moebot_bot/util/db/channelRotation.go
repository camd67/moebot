package db

import (
	"context"
	"log"
	"strings"

	"github.com/volatiletech/sqlboiler/boil"

	"github.com/volatiletech/sqlboiler/queries/qm"

	"github.com/volatiletech/sqlboiler/queries"

	"github.com/camd67/moebot/moebot_bot/util/db/models"
	"github.com/camd67/moebot/moebot_bot/util/db/types"
)

const (
	channelRotationQuery = `SELECT channel_rotation.operation_id as "operation.id", channel_rotation.current_channel_uid, 
							channel_rotation.channel_uids, scheduled_operation.server_id as "operation.server_id"
							FROM channel_rotation 
							INNER JOIN scheduled_operation ON scheduled_operation.id = channel_rotation.operation_id
							WHERE operation_id = $1`

	channelRotationInsert = `INSERT INTO channel_rotation (operation_id, current_channel_uid, channel_uids) VALUES($1, $2, $3)`
)

func ChannelRotationQuery(operationID int64) (*types.ChannelRotation, error) {
	var cr *types.ChannelRotation
	err := queries.Raw(channelRotationQuery, operationID).Bind(context.Background(), moeDb, cr)
	if err != nil {
		return nil, err
	}
	cr.ChannelUIDList = strings.Split(cr.ChannelUIDs, " ")
	return cr, nil
}

func ChannelRotationUpdate(operationID int64, currentChannelUID string) error {
	_, err := models.ChannelRotations(qm.Where("operation_id = ?", operationID)).
		UpdateAll(context.Background(), moeDb, models.M{"current_channel_uid": currentChannelUID})
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
	cr := &models.ChannelRotation{
		OperationID:       operation.ID,
		CurrentChannelUID: currentChannelUID,
		ChannelUids:       strings.Join(channels, " "),
	}
	err = cr.Insert(context.Background(), moeDb, boil.Infer())
	return err
}
