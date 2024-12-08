package planner

import (
	"time"
)

type Rule struct {
	Start    *TimeMark
	End      *TimeMark
	Services map[string]int
}

func (r *Rule) InPeriodNow() bool {
	now := time.Now()

	if (r.Start.Hour > 0 && now.Hour() < r.Start.Hour) ||
		(r.End.Hour > 0 && now.Hour() > r.End.Hour) {
		return false
	}

	if (r.Start.Weekday > 0 && int(now.Weekday()) < r.Start.Weekday) ||
		(r.End.Weekday > 0 && int(now.Weekday()) > r.End.Weekday) {
		return false
	}

	if (r.Start.Day > 0 && now.Day() < r.Start.Day) ||
		(r.End.Day > 0 && now.Day() > r.End.Day) {
		return false
	}

	if (r.Start.Month > 0 && int(now.Month()) < r.Start.Month) ||
		(r.End.Month > 0 && int(now.Month()) > r.End.Month) {
		return false
	}

	return true
}
