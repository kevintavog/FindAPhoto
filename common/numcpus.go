package common

import (
	"fmt"
	"runtime"
)

func NumCpus() int {
	return runtime.GOMAXPROCS(0)
}

func RatioNumCpus(factor float32) int {
	if factor <= 0 || factor > 1 {
		panic(fmt.Sprintf("'factor' must be > 0 & <= 1 (%f)", factor))
	}
	
	floatCpus := float32(NumCpus()) * factor
	cpus := int(floatCpus)
	extra := floatCpus - float32(cpus)
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
