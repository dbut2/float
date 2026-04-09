package utils

import (
	"time"
	_ "time/tzdata"
)

var Location = func() *time.Location {
	loc, err := time.LoadLocation("Australia/Melbourne")
	if err != nil {
		panic("failed to load Australia/Melbourne timezone: " + err.Error())
	}
	return loc
}()

func Now() time.Time {
	return time.Now().In(Location)
}

func Today() time.Time {
	n := Now()
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, Location)
}

func ToDate(t time.Time) time.Time {
	l := t.In(Location)
	return time.Date(l.Year(), l.Month(), l.Day(), 0, 0, 0, 0, Location)
}
