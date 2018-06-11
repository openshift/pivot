package utils

import (
	"fmt"
	"os"
)

// Fatal prints out a string to STDERR and then exists
func Fatal(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
