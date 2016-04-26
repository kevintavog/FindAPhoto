package common

import (
	"time"
)

const (
	janStart = 1
	febStart = 32
	marStart = 61
	aprStart = 92
	mayStart = 122
	junStart = 153
	julStart = 183
	augStart = 214
	sepStart = 245
	octStart = 275
	novStart = 306
	decStart = 336
)

var monthStarts = [...]int{janStart, febStart, marStart, aprStart, mayStart, junStart, julStart, augStart, sepStart, octStart, novStart, decStart}

func DayOfYearFromDate(date time.Time) int {
	return DayOfYear(int(date.Month()), date.Day())
}

func DayOfYear(month, day int) int {
	if month < 1 || month > 12 {
		return -1
	}

	if day < 1 || day > 31 {
		return -1
	}

	return monthStarts[month-1] + (day - 1)
}
