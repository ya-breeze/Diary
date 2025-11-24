package utils

import (
	"time"
)

type Granularity string

const (
	GranularityMonth Granularity = "month"
	GranularityYear  Granularity = "year"
)

func RoundToGranularity(date time.Time, granularity Granularity, roundUp bool,
) time.Time {
	add := 0
	if roundUp {
		add = 1
	}

	if granularity == GranularityMonth {
		return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location()).AddDate(0, add, 0)
	} else if granularity == GranularityYear {
		return time.Date(date.Year(), 1, 1, 0, 0, 0, 0, date.Location()).AddDate(add, 0, 0)
	}
	return date
}

func GetCurrentDate() string {
	return time.Now().Format("2006-01-02")
}
