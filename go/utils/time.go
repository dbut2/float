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
	return Now().Truncate(24 * time.Hour)
}

func ToDate(t time.Time) time.Time {
	return t.In(Location).Truncate(24 * time.Hour)
}
