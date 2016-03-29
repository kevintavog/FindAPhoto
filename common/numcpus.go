package common

import (
	"runtime"
)

func NumCpus() int {
	return runtime.GOMAXPROCS(0)
}

func RatioNumCpus(divisor int) int {
	numCpus := NumCpus() / divisor
	if numCpus < 1 {
		return 1
	}
	return numCpus
}
