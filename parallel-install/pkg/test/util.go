package test

import (
	"os"
	"path"
)

func GetTestDataDirectory() string {
	currentDir, err := os.Getwd()
	if err != nil {
		return ""
	}

	return path.Join(currentDir, "/../test/data")
}
