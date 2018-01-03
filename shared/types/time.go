package types

import (
	"time"
)

type TimeInterval interface {
	Includes(time.Time) bool
}

type RepeatingTimeInterval struct {
	Every     []time.Weekday `json:"every"`
	StartTime Clock          `json:"start_time"`
	EndTime   Clock          `json:"end_time"`
}

type Clock struct {
	Hour   int `json:"hour"`
	Minute int `json:"minute"`
}

type Moment interface {
	Weekday() time.Weekday
	Clock() (hour, minute, second int)
}

// Includes checks if the given time is in this RepeatingTimeInterval.
func (interval RepeatingTimeInterval) Includes(testTime Moment) bool {
	for _, weekday := range interval.Every {
		if testTime.Weekday() == weekday {
			testHour, testMinute, _ := testTime.Clock()
			if testHour < interval.StartTime.Hour {
				return false
			}
			if testHour < interval.StartTime.Hour || testHour > interval.EndTime.Hour {
				return false
			}
			if testHour == interval.StartTime.Hour {
				// Same as start hour, test minute.
				if testMinute < interval.StartTime.Minute {
					return false
				}
			}
			if testHour == interval.EndTime.Hour {
				// Same as end hour, test minute.
				if testMinute > interval.EndTime.Minute {
					return false
				}
			}
			return true
		}
	}
	return false
}
