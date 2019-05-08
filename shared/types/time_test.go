package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestTime struct {
	CurrentWeekday time.Weekday
	Hour           int
	Minute         int
}

func (t TestTime) Weekday() time.Weekday {
	return t.CurrentWeekday
}

func (t TestTime) Clock() (int, int, int) {
	return t.Hour, t.Minute, 0
}

func TestInSameStartHour(t *testing.T) {
	interval := RepeatingTimeInterval{
		[]time.Weekday{time.Sunday},
		Clock{0, 30},
		Clock{1, 0},
	}

	var testTime TestTime

	testTime = TestTime{time.Sunday, 0, 0}
	if interval.Includes(testTime) {
		t.Fatalf("%v was detected as including %v.",
			interval, testTime)
	}

	testTime = TestTime{time.Sunday, 0, 30}
	if !interval.Includes(testTime) {
		t.Fatalf("%v was not detected as including %v.",
			interval, testTime)
	}

	testTime = TestTime{time.Sunday, 0, 59}
	if !interval.Includes(testTime) {
		t.Fatalf("%v was not detected as including %v.",
			interval, testTime)
	}
}

func TestInSameEndHour(t *testing.T) {
	interval := RepeatingTimeInterval{
		[]time.Weekday{time.Sunday},
		Clock{0, 30},
		Clock{1, 0},
	}

	var testTime TestTime

	testTime = TestTime{time.Sunday, 1, 0}
	if !interval.Includes(testTime) {
		t.Fatalf("%v was not detected as including %v.",
			interval, testTime)
	}

	testTime = TestTime{time.Sunday, 1, 30}
	if interval.Includes(testTime) {
		t.Fatalf("%v was detected as including %v.",
			interval, testTime)
	}
}

func TestInBothHours(t *testing.T) {
	interval := RepeatingTimeInterval{
		[]time.Weekday{time.Sunday},
		Clock{2, 0},
		Clock{2, 30},
	}

	var testTime TestTime

	testTime = TestTime{time.Sunday, 2, 0}
	if !interval.Includes(testTime) {
		t.Fatalf("%v was not detected as including %v.",
			interval, testTime)
	}

	testTime = TestTime{time.Sunday, 2, 15}
	if !interval.Includes(testTime) {
		t.Fatalf("%v was not detected as including %v.",
			interval, testTime)
	}

	testTime = TestTime{time.Sunday, 2, 30}
	if !interval.Includes(testTime) {
		t.Fatalf("%v was not detected as including %v.",
			interval, testTime)
	}

	testTime = TestTime{time.Sunday, 2, 45}
	if interval.Includes(testTime) {
		t.Fatalf("%v was detected as including %v.",
			interval, testTime)
	}
}

func TestDifferentDay(t *testing.T) {
	interval := RepeatingTimeInterval{
		[]time.Weekday{time.Sunday},
		Clock{3, 0},
		Clock{3, 59},
	}

	var testTime TestTime

	testTime = TestTime{time.Monday, 3, 0}
	if interval.Includes(testTime) {
		t.Fatalf("%v was detected as including %v.",
			interval, testTime)
	}
}

func TestDifferentDayMatching(t *testing.T) {
	interval := RepeatingTimeInterval{
		[]time.Weekday{time.Sunday, time.Monday},
		Clock{3, 0},
		Clock{3, 59},
	}

	var testTime TestTime

	testTime = TestTime{time.Monday, 3, 0}
	if !interval.Includes(testTime) {
		t.Fatalf("%v was not detected as including %v.",
			interval, testTime)
	}
}

func TestWeekdayIntervalsSingle(t *testing.T) {
	intervals := RepeatingTimeIntervals{
		{
			[]time.Weekday{time.Monday},
			Clock{0, 0}, Clock{12, 0},
		},
	}

	// Only Monday should have an interval.
	weekdayToIntervals := intervals.weekdayIntervals()
	assert.Len(t, weekdayToIntervals[time.Sunday], 0)
	assert.Len(t, weekdayToIntervals[time.Monday], 1)
	assert.Len(t, weekdayToIntervals[time.Tuesday], 0)
	assert.Len(t, weekdayToIntervals[time.Wednesday], 0)
	assert.Len(t, weekdayToIntervals[time.Thursday], 0)
	assert.Len(t, weekdayToIntervals[time.Friday], 0)
	assert.Len(t, weekdayToIntervals[time.Saturday], 0)

	// Monday should be 0-12.
	assert.Equal(t, Interval{
		start: time.Date(0, 0, 0, 0, 0, 0, 0, time.Local),
		end:   time.Date(0, 0, 0, 12, 0, 0, 0, time.Local),
	}, weekdayToIntervals[time.Monday][0])
}

func TestWeekdayIntervalsMultiplePerDay(t *testing.T) {
	intervals := RepeatingTimeIntervals{
		{
			[]time.Weekday{time.Monday, time.Tuesday, time.Friday},
			Clock{3, 0}, Clock{6, 0},
		},
		{
			[]time.Weekday{time.Tuesday},
			Clock{4, 0}, Clock{7, 0},
		},
		{
			[]time.Weekday{time.Monday},
			Clock{8, 0}, Clock{10, 30},
		},
	}

	// Only Monday, Tuesday, and Friday should have intervals. Monday has 2.
	weekdayToIntervals := intervals.weekdayIntervals()
	assert.Len(t, weekdayToIntervals[time.Sunday], 0)
	assert.Len(t, weekdayToIntervals[time.Monday], 2)
	assert.Len(t, weekdayToIntervals[time.Tuesday], 1)
	assert.Len(t, weekdayToIntervals[time.Wednesday], 0)
	assert.Len(t, weekdayToIntervals[time.Thursday], 0)
	assert.Len(t, weekdayToIntervals[time.Friday], 1)
	assert.Len(t, weekdayToIntervals[time.Saturday], 0)

	// Monday should be 3-6 and 8-10:30.
	// In descending order of start time, so 8-10:30 first.
	assert.Equal(t, Interval{
		start: time.Date(0, 0, 0, 8, 0, 0, 0, time.Local),
		end:   time.Date(0, 0, 0, 10, 30, 0, 0, time.Local),
	}, weekdayToIntervals[time.Monday][0])
	assert.Equal(t, Interval{
		start: time.Date(0, 0, 0, 3, 0, 0, 0, time.Local),
		end:   time.Date(0, 0, 0, 6, 0, 0, 0, time.Local),
	}, weekdayToIntervals[time.Monday][1])

	// Tuesday should be 3-7.
	assert.Equal(t, Interval{
		start: time.Date(0, 0, 0, 3, 0, 0, 0, time.Local),
		end:   time.Date(0, 0, 0, 7, 0, 0, 0, time.Local),
	}, weekdayToIntervals[time.Tuesday][0])

	// Friday should be 3-6.
	assert.Equal(t, Interval{
		start: time.Date(0, 0, 0, 3, 0, 0, 0, time.Local),
		end:   time.Date(0, 0, 0, 6, 0, 0, 0, time.Local),
	}, weekdayToIntervals[time.Friday][0])
}

func TestWithDate(t *testing.T) {
	interval := Interval{
		start: time.Date(0, 0, 0, 3, 0, 0, 0, time.Local),
		end:   time.Date(0, 0, 0, 6, 0, 0, 0, time.Local),
	}

	var year int
	var month time.Month
	var day int

	var withDate Interval
	withDate = interval.WithDate(2019, 12, 31)
	year, month, day = withDate.start.Date()
	assert.Equal(t, 2019, year)
	assert.Equal(t, time.Month(12), month)
	assert.Equal(t, 31, day)
	year, month, day = withDate.end.Date()
	assert.Equal(t, 2019, year)
	assert.Equal(t, time.Month(12), month)
	assert.Equal(t, 31, day)

	// Should stay the same after doing WithDate again.
	withDate = interval.WithDate(2019, 12, 31)
	year, month, day = withDate.start.Date()
	assert.Equal(t, 2019, year)
	assert.Equal(t, time.Month(12), month)
	assert.Equal(t, 31, day)
	year, month, day = withDate.end.Date()
	assert.Equal(t, 2019, year)
	assert.Equal(t, time.Month(12), month)
	assert.Equal(t, 31, day)
}

func TestTimeOverlapStartEnd(t *testing.T) {
	// Test with an interval from 3-6:30 and 10-12.
	intervals := []Interval{
		{
			start: time.Date(0, 0, 0, 3, 0, 0, 0, time.Local),
			end:   time.Date(0, 0, 0, 6, 30, 0, 0, time.Local),
		},
		{
			start: time.Date(0, 0, 0, 10, 0, 0, 0, time.Local),
			end:   time.Date(0, 0, 0, 12, 0, 0, 0, time.Local),
		},
	}

	// Get the overlap between those intervals and 4-11.
	// Start and end must be at the same date. Let's try Dec 31st, 2019.
	start := time.Date(2019, 12, 31, 4, 0, 0, 0, time.Local)
	end := time.Date(2019, 12, 31, 11, 0, 0, 0, time.Local)

	overlap := timeOverlap(intervals, &start, &end)
	// The expected overlap is 2.5 hours for the first interval,
	// and 1 hour for the second interval, so 3.5 hours.
	assert.Equal(t, 3.5, overlap.Hours())
}

func TestTimeOverlapStart(t *testing.T) {
	// Test with an interval from 3-6:30 and 10-12.
	intervals := []Interval{
		{
			start: time.Date(0, 0, 0, 3, 0, 0, 0, time.Local),
			end:   time.Date(0, 0, 0, 6, 30, 0, 0, time.Local),
		},
		{
			start: time.Date(0, 0, 0, 10, 0, 0, 0, time.Local),
			end:   time.Date(0, 0, 0, 12, 0, 0, 0, time.Local),
		},
	}

	// Get the overlap between those intervals and 1 to end of day.
	// Let's try Dec 31st, 2019.
	start := time.Date(2019, 12, 31, 1, 0, 0, 0, time.Local)

	overlap := timeOverlap(intervals, &start, nil)
	// The expected overlap is 3.5 hours for the first interval,
	// and 2 hours for the second interval, so 5.5 hours.
	assert.Equal(t, 5.5, overlap.Hours())
}

func TestTimeOverlapEnd(t *testing.T) {
	// Test with an interval from 3-6:30 and 10-12.
	intervals := []Interval{
		{
			start: time.Date(0, 0, 0, 3, 0, 0, 0, time.Local),
			end:   time.Date(0, 0, 0, 6, 30, 0, 0, time.Local),
		},
		{
			start: time.Date(0, 0, 0, 10, 0, 0, 0, time.Local),
			end:   time.Date(0, 0, 0, 12, 0, 0, 0, time.Local),
		},
	}

	// Get the overlap between those intervals and start of day to 11.
	// Let's try Dec 31st, 2019.
	end := time.Date(2019, 12, 31, 11, 0, 0, 0, time.Local)

	overlap := timeOverlap(intervals, nil, &end)
	// The expected overlap is 3.5 hours for the first interval,
	// and 1 hour for the second interval, so 4.5 hours.
	assert.Equal(t, 4.5, overlap.Hours())
}

func TestTotalOverlapSingleDayNoIntervals(t *testing.T) {
	intervals := RepeatingTimeIntervals{}

	// Let's start from Dec 31st, 2019, so it rolls over a year.
	// Dec 31st, 2019 is a Tuesday.
	start := time.Date(2019, 12, 31, 0, 0, 0, 0, time.Local)
	end := time.Date(2019, 12, 31, 24, 0, 0, 0, time.Local)

	// With no intervals, the overlap should be 0.
	overlap := intervals.TotalOverlap(start, end)
	assert.Equal(t, 0.0, overlap.Hours())
}

func TestTotalOverlapSingleDayWrongDayOfWeek(t *testing.T) {
	// 0-3, 6-9 on a Tuesday.
	intervals := RepeatingTimeIntervals{
		{
			[]time.Weekday{time.Monday},
			Clock{0, 0}, Clock{3, 0},
		},
		{
			[]time.Weekday{time.Monday},
			Clock{6, 0}, Clock{9, 0},
		},
	}

	// Let's start from Dec 31st, 2019, so it rolls over a year.
	// Dec 31st, 2019 is a Tuesday.
	start := time.Date(2019, 12, 31, 0, 0, 0, 0, time.Local)
	end := time.Date(2019, 12, 31, 24, 0, 0, 0, time.Local)

	// With no intervals on Tuesday, the total should be 0.
	overlap := intervals.TotalOverlap(start, end)
	assert.Equal(t, 0.0, overlap.Hours())
}

func TestTotalOverlapSingleDay(t *testing.T) {
	// 0-3, 6-9 on a Tuesday.
	intervals := RepeatingTimeIntervals{
		{
			[]time.Weekday{time.Tuesday},
			Clock{0, 0}, Clock{3, 0},
		},
		{
			[]time.Weekday{time.Tuesday},
			Clock{6, 0}, Clock{9, 0},
		},
	}

	// Let's start from Dec 31st, 2019, so it rolls over a year.
	// Dec 31st, 2019 is a Tuesday.
	start := time.Date(2019, 12, 31, 0, 0, 0, 0, time.Local)
	end := time.Date(2019, 12, 31, 24, 0, 0, 0, time.Local)

	// With 0-3, 6-9 intervals, the overlap should be 6 hours total.
	overlap := intervals.TotalOverlap(start, end)
	assert.Equal(t, 6.0, overlap.Hours())
}

func TestTotalOverlapSingleDayPartial(t *testing.T) {
	// 0-3, 6-9 on a Tuesday.
	intervals := RepeatingTimeIntervals{
		{
			[]time.Weekday{time.Tuesday},
			Clock{0, 0}, Clock{3, 0},
		},
		{
			[]time.Weekday{time.Tuesday},
			Clock{6, 0}, Clock{9, 0},
		},
	}

	// Let's start from Dec 31st, 2019, so it rolls over a year.
	// Dec 31st, 2019 is a Tuesday.
	// Let's start from 2:15 and end at 7.
	start := time.Date(2019, 12, 31, 2, 15, 0, 0, time.Local)
	end := time.Date(2019, 12, 31, 7, 0, 0, 0, time.Local)

	// With 0-3, 6-9 intervals, the overlap should be 45 minutes + 1 hour,
	// for 1 hour 45 minutes total.
	overlap := intervals.TotalOverlap(start, end)
	assert.Equal(t, 1.75, overlap.Hours())
}

func TestTotalOverlapTwoDay(t *testing.T) {
	// 0-3, 6-9 on a Tuesday, and 12-20:30 on Wednesday.
	intervals := RepeatingTimeIntervals{
		{
			[]time.Weekday{time.Tuesday},
			Clock{0, 0}, Clock{3, 0},
		},
		{
			[]time.Weekday{time.Tuesday},
			Clock{6, 0}, Clock{9, 0},
		},
		{
			[]time.Weekday{time.Wednesday},
			Clock{12, 0}, Clock{20, 30},
		},
	}

	// Let's start from Dec 31st, 2019, so it rolls over a year.
	// Dec 31st, 2019 is a Tuesday.
	start := time.Date(2019, 12, 31, 0, 0, 0, 0, time.Local)
	// End goes 48 hours into the future.
	end := time.Date(2019, 12, 31, 48, 0, 0, 0, time.Local)

	// With 0-3, 6-9 intervals, the overlap should be 6 hours total on Tuesday.
	// With the 12-20:30 interval, the overlap should be 8.5 hours on Wednesday.
	// The total overlap should be 14.5 hours.
	overlap := intervals.TotalOverlap(start, end)
	assert.Equal(t, 14.5, overlap.Hours())
}

func TestTotalOverlapTwoDayPartial(t *testing.T) {
	// 0-3, 6-9 on a Tuesday, and 12-20:30 on Wednesday.
	intervals := RepeatingTimeIntervals{
		{
			[]time.Weekday{time.Tuesday},
			Clock{0, 0}, Clock{3, 0},
		},
		{
			[]time.Weekday{time.Tuesday},
			Clock{6, 0}, Clock{9, 0},
		},
		{
			[]time.Weekday{time.Wednesday},
			Clock{12, 0}, Clock{20, 30},
		},
	}

	// Let's start from Dec 31st, 2019, so it rolls over a year.
	// Dec 31st, 2019 is a Tuesday.
	// Let's start at 2 (2am) on Tuesday.
	start := time.Date(2019, 12, 31, 2, 0, 0, 0, time.Local)
	// Let's end at 17 (5pm) on Wednesday.
	end := time.Date(2020, 1, 1, 17, 0, 0, 0, time.Local)

	// With 0-3, 6-9 intervals, the overlap should be 4 hours total on Tuesday.
	// With the 12-20:30 interval, the overlap should be 5 hours on Wednesday.
	// The total overlap should be 9 hours.
	overlap := intervals.TotalOverlap(start, end)
	assert.Equal(t, 9.0, overlap.Hours())
}

func TestTotalOverlapMultiday(t *testing.T) {
	// 0-3, 6-9 on a Tuesday, 10-22 on Wednesday, 0-12 on Thursday,
	// and 12-20:30 on Friday.
	intervals := RepeatingTimeIntervals{
		{
			[]time.Weekday{time.Tuesday},
			Clock{0, 0}, Clock{3, 0},
		},
		{
			[]time.Weekday{time.Tuesday},
			Clock{6, 0}, Clock{9, 0},
		},
		{
			[]time.Weekday{time.Wednesday},
			Clock{10, 0}, Clock{22, 0},
		},
		{
			[]time.Weekday{time.Thursday},
			Clock{0, 0}, Clock{12, 0},
		},
		{
			[]time.Weekday{time.Friday},
			Clock{12, 0}, Clock{20, 30},
		},
	}

	// Let's start from Dec 31st, 2019, so it rolls over a year.
	// Dec 31st, 2019 is a Tuesday.
	// Let's start at 2 (2am) on Tuesday.
	start := time.Date(2019, 12, 31, 2, 0, 0, 0, time.Local)
	// Let's end at 17 (5pm) on Friday.
	end := time.Date(2020, 1, 3, 17, 0, 0, 0, time.Local)

	// With 0-3, 6-9 intervals, the overlap should be 4 hours total on Tuesday.
	// Wednesday and Thursday both have 12 hour intervals.
	// With the 12-20:30 interval, the overlap should be 5 hours on Friday.
	// The total overlap should be 33 hours.
	overlap := intervals.TotalOverlap(start, end)
	assert.Equal(t, 33.0, overlap.Hours())
}

func TestTotalOverlapMultimonth(t *testing.T) {
	// 0-3, 6-9 on a Tuesday, 10-22 on Wednesday, 0-12 on Thursday,
	// and 12-20:30 on Friday.
	intervals := RepeatingTimeIntervals{
		{
			[]time.Weekday{time.Tuesday},
			Clock{0, 0}, Clock{3, 0},
		},
		{
			[]time.Weekday{time.Tuesday},
			Clock{6, 0}, Clock{9, 0},
		},
		{
			[]time.Weekday{time.Wednesday},
			Clock{10, 0}, Clock{22, 0},
		},
		{
			[]time.Weekday{time.Thursday},
			Clock{0, 0}, Clock{12, 0},
		},
		{
			[]time.Weekday{time.Friday},
			Clock{12, 0}, Clock{20, 30},
		},
	}

	// Let's start from Dec 31st, 2019, so it rolls over a year.
	// Dec 31st, 2019 is a Tuesday.
	// Let's start at 2 (2am) on Tuesday.
	start := time.Date(2019, 12, 31, 2, 0, 0, 0, time.Local)
	// Let's end at 17 (5pm) on Friday, 10 weeks later.
	end := time.Date(2020, 3, 6, 17, 0, 0, 0, time.Local)

	// The first Tuesday has a 4 hour overlap.
	// The last Friday has a 5 hour overlap.
	// Other Tuesdays have a 6 hour overlap.
	// Wednesday and Thursdays have 12 hour overlaps.
	// Other Fridays has a 8.5 hour overlap.

	// Between 2019-12-31 and 2020-3-3, there are:
	//  10 Tuesdays (4 hours + 9 * 6 hours).
	//  10 Wednesdays (10 * 12 hours)
	//  10 Thursdays (10 * 12 hours)
	//  10 Fridays (9 * 8.5 hours + 5 hours)
	// Total: 379.5 hours.

	overlap := intervals.TotalOverlap(start, end)
	assert.Equal(t, 379.5, overlap.Hours())
}
