package common

import (
	"os/exec"
)

func IsExecWorking(commandName string, args ...string) bool {
	out, err := CheckExec(commandName, args...)
	return err == nil && len(out) > 0
}

func CheckExec(commandName string, args ...string) (string, error) {
	b, err := exec.Command(commandName, args...).Output()
	return string(b), err
}
