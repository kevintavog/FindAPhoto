package main

import (
	"fmt"
)

var majorVersion = 1
var minorVersion = 1
var buildVersion = 13

func versionString() string {
	return fmt.Sprintf("%d.%d.%d", majorVersion, minorVersion, buildVersion)
}
