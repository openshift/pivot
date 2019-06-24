package utils

import (
	"os"

	"github.com/golang/glog"
)

// FileExists checks if the file exists, gracefully handling ENOENT.
func FileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	glog.Warningf("Failed to stat %s: %v", path, err)
	return false
}
