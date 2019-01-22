package utils

import (
	"testing"
)

// TestRun should always pass. The function will panic if it is unable to
// execute the shell command(s) or the command returns non-zero.
func TestRun(t *testing.T) {
	Run("echo", "echo", "from", "TestRun")
}

// TestRunGetOut verifies the output of running a command is
// its output, trimmed of whitespace.
func TestRunGetOut(t *testing.T) {
	if result := RunGetOut("echo", "hello", "world"); result != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", result)
	}
}
