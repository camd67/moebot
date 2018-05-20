package util

import "time"

type Timer struct {
	Marks []TimerMark
}

type TimerMark struct {
	mark time.Time // Start time for this timer mark. Internal since it doesn't mean much to an end user

	Duration time.Duration `json:"d"` // Duration between this mark and the next one. Will not be populated until the timer is stopped
	Name     string        `json:"n"` // Name of the marker
}

const (
	// Timer marker key for the total time elapsed in the timer
	TimerMarkTotal = "_total"
	// Timer marker key for the first event in the timer. This will always be present
	TimerMarkStart = "_start"
	// Timer marker key for the last event in the timer. This will always be present
	TimerMarkEnd = "_end"
	// Timer marker keys for commands
	TimerMarkCommandBegin = "command_begin_"
	TimerMarkCommandEnd   = "command_end_"
	// Timer marker keys for db interactions
	TimerMarkDbBegin = "db_begin_"
	TimerMarkDbEnd   = "db_end_"
)

/*
Creates a new Timer object with an initial "Start" time.
As soon as you call this the timer will start
*/
func StartNamedTimer(firstName string) Timer {
	firstMark := TimerMark{
		mark: time.Now(),
	}
	if firstName == "" {
		firstMark.Name = TimerMarkStart
	} else {
		firstMark.Name = firstName
	}
	// Always start with a single timer
	return Timer{
		Marks: []TimerMark{firstMark},
	}
}

/*
Creates a new timer with the default name
*/
func StartTimer() Timer {
	return Timer{
		Marks: []TimerMark{{
			mark: time.Now(),
			Name: TimerMarkStart,
		}},
	}
}

/*
Adds a single mark with a name. This should be called like the following:

func main() {
	t := StartNamedTimer()
	t.AddMark("Processing Names")
	processNames()
	t.StopTimer()
}

This will give you a marker for "Processing Names" that was the total time processNames took
*/
func (t *Timer) AddMark(name string) {
	t.Marks = append(t.Marks, TimerMark{
		mark: time.Now(),
		Name: name,
	})
}

/*
Stops this timer, returning back a map of names to durations.
*/
func (t *Timer) StopTimer() map[string]time.Duration {
	// ALWAYS stop the timer first before doing anything else
	t.AddMark(TimerMarkEnd)

	times := make(map[string]time.Duration, len(t.Marks))

	// add a special extra mark for total
	totalDuration := t.Marks[len(t.Marks)-1].mark.Sub(t.Marks[0].mark)
	times[TimerMarkTotal] = totalDuration
	t.Marks = append(t.Marks, TimerMark{
		mark:     time.Now(),
		Duration: totalDuration,
		Name:     TimerMarkTotal,
	})

	// Then go over each mark and figure out the difference between the current one and the previous
	for i := 1; i < len(t.Marks); i++ {
		currentDuration := t.Marks[i].mark.Sub(t.Marks[i-1].mark)
		times[t.Marks[i-1].Name] = currentDuration
		t.Marks[i-1].Duration = currentDuration
	}
	// Remove the last mark since it's no longer needed
	t.Marks = t.Marks[:len(t.Marks)-1]
	// Don't add an extra duration for the stop event
	return times
}
