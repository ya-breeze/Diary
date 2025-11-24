package utils

import (
	"time"
)

func FormatTime(t time.Time, format string) string {
	return t.Format(format)
}
