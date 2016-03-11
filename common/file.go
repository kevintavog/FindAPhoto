package common

import (
	"os"
)

func CreateDirectory(directory string) error {
	return os.MkdirAll(directory, os.ModePerm)
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	// An error other that the file doesn't exist
	if err != nil {
		return false, err
	}
	return true, nil
}

func FileExists(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	// An error other that the file doesn't exist
	if err != nil {
		return false, err
	}
	return !fileInfo.IsDir(), nil
}
