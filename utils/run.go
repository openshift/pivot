package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

// RunGetOutln executes a system command, turns the result into a string
// and trims white space before and after the string
func RunGetOutln(command string, args ...string) string {
	rawOut := Run(command, args...)
	return strings.TrimSpace(string(rawOut))
}

// Run executes a command on the local system and returns the output
// in string format
func Run(command string, args ...string) string {
	fmt.Printf("Running: %s %s\n", command, strings.Join(args, " "))
	rawOut, err := exec.Command(command, args...).Output()
	if err != nil {
		Fatal(fmt.Sprintf("Unable to run command %s: %s", command, err))
	}
	return string(rawOut)
}
