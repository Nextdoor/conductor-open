package types

import (
	"testing"
	"time"
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
