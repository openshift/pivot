package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

// runGetOutln executes a system command, turns the result into a string
// and trims white space before and after the string
func RunGetOutln(command string, args ...string) string {
	rawOut := Run(command, args...)
	return strings.TrimSpace(string(rawOut))
}

// run executes a command on the local system and returns the output
// in string format
func Run(command string, args ...string) string {
	rawOut, err := exec.Command(command, args...).Output()
	if err != nil {
		fatal(fmt.Sprintf("Unable to run command %s: %s", command, err))
	}
	return string(rawOut)
}
