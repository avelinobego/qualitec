package util

import "time"

func FirstDate(t time.Time) (result time.Time) {
	result = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	return
}

func LastDate(t time.Time) (result time.Time) {
	result = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.Local)
	return
}
