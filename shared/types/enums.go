package types

import (
	"fmt"
)

type Mode int

const (
	Schedule Mode = iota
	Manual
)

func (m Mode) String() string {
	switch m {
	case Schedule:
		return "schedule"
	case Manual:
		return "manual"
	default:
		panic(fmt.Errorf("Unknown mode: %f", m))
	}
}

func (m Mode) IsScheduleMode() bool {
	return m == Schedule
}

func (m Mode) IsManualMode() bool {
	return m == Manual
}

func ModeFromString(mode string) (Mode, error) {
	switch mode {
	case Schedule.String():
		return Schedule, nil
	case Manual.String():
		return Manual, nil
	default:
		return Schedule, fmt.Errorf("Unknown mode %s", mode)
	}
}

type PhaseType int

const (
	Delivery PhaseType = iota
	Verification
	Deploy
)

func (e PhaseType) String() string {
	switch e {
	case Delivery:
		return "delivery"
	case Verification:
		return "verification"
	case Deploy:
		return "deploy"
	default:
		panic(fmt.Errorf("Unknown mode: %d", e))
	}
}

func PhaseTypeFromString(phaseType string) (PhaseType, error) {
	switch phaseType {
	case "delivery":
		return Delivery, nil
	case "verification":
		return Verification, nil
	case "deploy":
		return Deploy, nil
	default:
		return -1, fmt.Errorf("Unknown phase type: %s", phaseType)
	}
}

type JobResult int

const (
	Ok JobResult = iota
	Error
)

func (j JobResult) String() string {
	switch j {
	case Ok:
		return "ok"
	case Error:
		return "error"
	default:
		panic(fmt.Errorf("Unknown mode: %d", j))
	}
}

func (j JobResult) IsValid() bool {
	return j >= Ok && j <= Error
}
