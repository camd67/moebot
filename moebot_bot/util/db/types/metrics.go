package types

import (
	"github.com/camd67/moebot/moebot_bot/util/event"
)

type MetricType int16

type MetricTimerJson struct {
	Events []event.TimerMark `json:"events"`
	UserId int               `json:"userId"`
}
