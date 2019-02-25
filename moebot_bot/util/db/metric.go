package db

import (
	"context"
	"encoding/json"
	"log"

	"github.com/volatiletech/sqlboiler/boil"

	"github.com/camd67/moebot/moebot_bot/util/db/models"
	"github.com/camd67/moebot/moebot_bot/util/db/types"
	"github.com/camd67/moebot/moebot_bot/util/event"
)

const (
	/*
		Metric representing a timer. This should store JSON data regarding timers and time data
	*/
	MetricTypeTimer types.MetricType = 1
)

func MetricInsertTimer(metric event.Timer, user models.UserProfile) error {
	jsonData, err := json.Marshal(types.MetricTimerJson{
		Events: metric.Marks,
		UserId: user.ID,
	})
	if err != nil {
		log.Println("Failed to serialize JSON data for metric timer", err)
		return err
	}
	m := &models.Metric{
		MetricType: MetricTypeTimer,
		Data:       jsonData,
	}
	err = m.Insert(context.Background(), moeDb, boil.Infer())
	if err != nil {
		log.Println("Failed to write to metric table", err)
	}
	return err
}
