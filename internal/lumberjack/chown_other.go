//go:build !linux
// +build !linux

package lumberjack

import (
	"os"
)

func chown(string, os.FileInfo) error {
	return nil
}
