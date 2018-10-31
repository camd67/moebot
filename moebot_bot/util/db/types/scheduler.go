package types

import "time"

type SchedulerType int

type ScheduledOperation struct {
	ID                   int64
	ServerID             int
	Type                 SchedulerType
	PlannedExecutionTime time.Time
}

type ChannelRotation struct {
	ChannelUIDList    []string
	CurrentChannelUID string
	ScheduledOperation
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
