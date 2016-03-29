package common

import (
	"math"
	"runtime"
)

func NumCpus() int {
	return runtime.GOMAXPROCS(0)
}

func RatioNumCpus(divisor int) int {
	cpus := NumCpus() / divisor
	extra := math.Mod(float64(NumCpus()), float64(divisor))
	if extra > 0.1 {
		cpus += 1
	}

	if cpus < 1 {
		return 1
	}
	return cpus
}

func MaxCpus(desired int) int {
	numCpus := NumCpus()
	if numCpus < desired {
		return numCpus
	}
	return desired
}
