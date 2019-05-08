package types

import (
	"sort"
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

type RepeatingTimeIntervals []RepeatingTimeInterval

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

type Interval struct {
	start time.Time
	end   time.Time
}

func (interval Interval) WithDate(year int, month time.Month, day int) Interval {
	return Interval{
		start: time.Date(
			year, month, day,
			interval.start.Hour(), interval.start.Minute(),
			0, 0, interval.start.Location()),
		end: time.Date(
			year, month, day,
			interval.end.Hour(), interval.end.Minute(),
			0, 0, interval.end.Location()),
	}
}

type Intervals []Interval

func (intervals Intervals) Len() int {
	return len(intervals)
}
func (intervals Intervals) Swap(i, j int) {
	intervals[i], intervals[j] = intervals[j], intervals[i]
}
func (intervals Intervals) Less(i, j int) bool {
	return intervals[i].start.Sub(intervals[j].start) < 0
}

func (repeatingTimeIntervals RepeatingTimeIntervals) weekdayIntervals() map[time.Weekday][]Interval {
	// Start by calculating the minimal set of intervals for the weekdays.
	weekdayToInterval := make(map[time.Weekday][]Interval)
	for weekday := time.Sunday; weekday <= time.Saturday; weekday++ {
		intervals := make([]Interval, 0)
		for _, interval := range repeatingTimeIntervals {
			includesWeekday := false
			for _, every := range interval.Every {
				if every == weekday {
					includesWeekday = true
				}
			}
			if !includesWeekday {
				continue
			}

			intervals = append(intervals, Interval{
				start: time.Date(0, 0, 0, interval.StartTime.Hour, interval.StartTime.Minute, 0, 0, time.Local),
				end:   time.Date(0, 0, 0, interval.EndTime.Hour, interval.EndTime.Minute, 0, 0, time.Local),
			})
		}

		// Sort and merge intervals for the weekday.
		// Sort by descending order of start time.
		sort.Sort(sort.Reverse(Intervals(intervals)))

		// Merge intervals as much as possible.
		mergedIntervals := make([]Interval, 0)
		mergedIntervalIndex := 0
		for i, interval := range intervals {
			if i == 0 {
				mergedIntervals = append(mergedIntervals, interval)
			} else {
				previousInterval := mergedIntervals[i-1]
				doesOverlap := interval.end.Sub(previousInterval.start) > 0
				if doesOverlap {
					previousEndsEarlier := interval.end.Sub(previousInterval.end) > 0
					if previousEndsEarlier {
						mergedIntervals[i-1] = interval
					} else {
						mergedIntervals[i-1] = Interval{
							start: interval.start,
							end:   previousInterval.end,
						}
					}
				} else {
					mergedIntervals = append(mergedIntervals, interval)
					mergedIntervalIndex += 1
				}
			}
		}

		// Store this set of intervals on the weekday.
		weekdayToInterval[weekday] = mergedIntervals
	}

	return weekdayToInterval
}

// Computes the time overlap between a specified array of intervals and
// a specified start and end time.
// Intervals must be sorted in descending order of start time.
// Start and end times must be on the same date.
// Start and end time are optional.
// If start time is nil, start time is assumed the very beginning of the day.
// If end time is nil, end time is assumed the very end of the day.
func timeOverlap(intervals []Interval, start *time.Time, end *time.Time) time.Duration {
	var year int
	var month time.Month
	var day int

	if start != nil {
		year, month, day = start.Date()
	} else if end != nil {
		year, month, day = end.Date()
	}

	var totalOverlap time.Duration = 0
	for _, interval := range intervals {
		datedInterval := interval.WithDate(year, month, day)
		testStart := datedInterval.start
		if start != nil && datedInterval.start.Sub(*start) < 0 {
			testStart = *start
		}
		testEnd := datedInterval.end
		if end != nil && datedInterval.end.Sub(*end) > 0 {
			testEnd = *end
		}
		overlap := testEnd.Sub(testStart)
		if overlap > 0 {
			totalOverlap += overlap
		}
	}

	return totalOverlap
}

// TotalOverlap calculates the total overlap duration between the specified start and end times.
func (repeatingTimeIntervals RepeatingTimeIntervals) TotalOverlap(start time.Time, end time.Time) time.Duration {
	if end.Sub(start) < 0 {
		// End is before start, so no overlap.
		return 0
	}

	weekdayToInterval := repeatingTimeIntervals.weekdayIntervals()
	// Iterate through each day between start and end time and calculate totalOverlap, summing it all up.
	//
	// If the start time and the end time are the same date, the totalOverlap is:
	//   `min(interval_end, end_time) - max(interval_start, start_time)` summed for each interval.
	//
	// Otherwise, the totalOverlap happens over multiple days.
	// For the first day, the totalOverlap is `interval_end - max(interval_start, start_time)` summed for each interval.
	// For the end day, the totalOverlap is `min(interval_end, end_time) - interval_start` summed for each interval.
	// For all middle days, the totalOverlap is `interval_end - interval_start` summed for each interval.
	var totalOverlap time.Duration = 0

	startYear, startMonth, startDay := start.Date()
	endYear, endMonth, endDay := end.Date()
	if startYear == endYear && startMonth == endMonth && startDay == endDay {
		// Start and end times are on the same date.
		// Find the relevant day intervals.
		totalOverlap += timeOverlap(weekdayToInterval[start.Weekday()], &start, &end)
	} else {
		// Start day.
		totalOverlap += timeOverlap(weekdayToInterval[start.Weekday()], &start, nil)

		// Middle days. Start one day ahead of start day, iterate through each day.
		middleTime := time.Date(startYear, startMonth, startDay, 0, 0, 0, 0, start.Location())
		middleTime = middleTime.AddDate(0, 0, 1)
		for {
			middleYear, middleMonth, middleDay := middleTime.Date()
			if middleYear == endYear && middleMonth == endMonth && middleDay == endDay {
				// On the end day.
				break
			}

			totalOverlap += timeOverlap(weekdayToInterval[middleTime.Weekday()], nil, nil)

			// Go forward one day.
			middleTime = middleTime.AddDate(0, 0, 1)
		}

		// End day.
		totalOverlap += timeOverlap(weekdayToInterval[end.Weekday()], nil, &end)
	}

	return totalOverlap
}
