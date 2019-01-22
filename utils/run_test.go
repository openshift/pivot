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

// TestRunExt verifies that the wait machinery works, even though we're only
// just testing a single step here since it's tricky to test retries.
func TestRunExt(t *testing.T) {
	RunExt(false, 0, "echo", "echo", "from", "TestRunExt")

	if result := RunExt(true, 0, "echo", "hello", "world"); result != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", result)
	}

	Run("echo", "This may take a while and then fail if you're really unlucky...")
	RunExt(false, 5, "sh", "-c", "[ $(($RANDOM % 100)) -lt 50 ]")
}
