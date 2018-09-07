package utils

import (
	"testing"
)

// TestRun should always pass. The function will panic if it is unable to
// execute the shell command(s) or the command returns non-zero.
func TestRun(t *testing.T) {
	Run("echo", "a")
}

// TestRunGetOut verifies the output of running a shell command is
// a string ending in a new line.
func TestRunGetOut(t *testing.T) {
	if result := RunGetOut("echo", "a"); result != "a\n" {
		t.Errorf("expected 'a\n', got '%s'", result)
	}
}

// TestRunGetOutln verifies the output of running a shell command is
// a string NOT ending in a new line.
func TestRunGetOutln(t *testing.T) {
	if result := RunGetOutln("echo", "a"); result != "a" {
		t.Errorf("expected 'a', got '%s'", result)
	}
}
