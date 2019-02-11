package util

import (
	"fmt"
	"strconv"
)

func IntFromString(name string, contents string) int {
	v, err := strconv.Atoi(contents)
	if err != nil {
		panic(&InvalidRequest{Message: fmt.Sprintf("'%s' is not an int: %s", name, contents)})
	}
	return v
}
