// clarifaiProvider_test.go
package main

import (
	"fmt"
	"testing"
)

func TestThreshold(t *testing.T) {

	for i := 0; i < 30; i++ {
		fmt.Printf("Doing loop %d \n", i+1)
		fakeClassify("")

	}
}
