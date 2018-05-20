package db

import (
	"encoding/json"
	"log"

	"github.com/camd67/moebot/moebot_bot/util"
)

type MetricType int

const (
	/*
		Metric representing a timer. This should store JSON data regarding timers and time data
	*/
	MetricTypeTimer MetricType = 1
)

type MetricTimerJson struct {
	Events []util.TimerMark `json:"events"`
	UserId int              `json:"userId"`
}

type Metric struct {
	Id   int
	Type MetricType
	Data json.RawMessage
}

const (
	metricTable = `CREATE TABLE IF NOT EXISTS metric(
		Id SERIAL NOT NULL PRIMARY KEY,
		Type SMALLINT NOT NULL,
		Data jsonb NOT NULL
	)`

	metricInsert = `INSERT INTO metric(Type, Data) VALUES ($1, $2)`
)

func MetricInsertTimer(metric util.Timer, user UserProfile) error {
	jsonData, err := json.Marshal(MetricTimerJson{
		Events: metric.Marks,
		UserId: user.Id,
	})
	if err != nil {
		log.Println("Failed to serialize JSON data for metric timer", err)
		return err
	}
	_, err = moeDb.Exec(metricInsert, MetricTypeTimer, jsonData)
	if err != nil {
		log.Println("Failed to write to metric table", err)
	}
	return err
}

func metricCreateTable() {
	_, err := moeDb.Exec(metricTable)
	if err != nil {
		log.Println("Error creating metric table", err)
		return
	}
}
