package utils

import "time"

// FormatDate formats a time.Time as "M/D/YYYY"
func FormatDate(t time.Time) string {
	return t.Format("1/2/2006")
}
