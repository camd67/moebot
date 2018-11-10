package types

import (
	"encoding/json"

	"github.com/camd67/moebot/moebot_bot/util/event"
)

type MetricType int

type MetricTimerJson struct {
	Events []event.TimerMark `json:"events"`
	UserId int               `json:"userId"`
}

type Metric struct {
	Id   int
	Type MetricType
	Data json.RawMessage
}
