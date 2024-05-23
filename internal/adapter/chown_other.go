//go:build !linux
// +build !linux

package adapter

import (
	"os"
)

func chown(string, os.FileInfo) error {
	return nil
}
