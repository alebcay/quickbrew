package utils

import (
	"os"
)

func IsFileReadable(filename string) bool {
	_, err := os.Open(filename)
	return err == nil
}

func IsFileExecutable(filename string) bool {
	info, err := os.Lstat(filename)
	if err != nil { return false }
	return info.Mode() & 0111 != 0
}

