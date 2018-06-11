package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Execute a command, logging it, and exit with a fatal error if
// the command failed.
func Run(command string, args ...string) {
	fmt.Printf("Running: %s %s\n", command, strings.Join(args, " "));
	cmd := exec.Command(command, args...);
	// Pass through by default
	cmd.Stdout = os.Stdout;
	cmd.Stderr = os.Stderr;
	err := cmd.Run()
	if err != nil {
		Fatal(fmt.Sprintf("%s: %s", command, err))
	}
}

// Like Run(), but get the output as a string
func RunGetOut(command string, args ...string) string {
	fmt.Printf("Running: %s %s\n", command, strings.Join(args, " "));
	cmd := exec.Command(command, args...);
	// Pass through by default
	cmd.Stderr = os.Stderr;
	rawOut, err := cmd.Output()
	if err != nil {
		Fatal(fmt.Sprintf("%s: %s", command, err))
	}
	return string(rawOut)
}

// Run executes a command on the local system and returns the output
// in string format
func RunGetOutln(command string, args ...string) string {
	return strings.TrimSpace(RunGetOut(command, args...));
}
