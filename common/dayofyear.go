package common

import (
	"time"
)

const (
	janStart = 1
	febStart = janStart + 31
	marStart = febStart + 29
	aprStart = marStart + 31
	mayStart = aprStart + 30
	junStart = mayStart + 31
	julStart = junStart + 30
	augStart = julStart + 31
	sepStart = augStart + 31
	octStart = sepStart + 30
	novStart = octStart + 31
	decStart = novStart + 30
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
