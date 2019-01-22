package utils

import (
	"bytes"
	"os"
	"os/exec"
	"strings"

	"github.com/golang/glog"
)

func runImpl(capture bool, command string, args ...string) ([]byte, error) {
	glog.Infof("Running: %s %s\n", command, strings.Join(args, " "))
	cmd := exec.Command(command, args...)
	cmd.Stderr = os.Stderr
	var stdout bytes.Buffer
	if !capture {
		cmd.Stdout = os.Stdout
	} else {
		cmd.Stdout = &stdout
	}
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	if capture {
		return stdout.Bytes(), nil
	}
	return []byte{}, nil
}

// Execute a command, logging it, and exit with a fatal error if
// the command failed.
func Run(command string, args ...string) {
	if _, err := runImpl(false, command, args...); err != nil {
		glog.Fatalf("%s: %s", command, err)
	}
}

// Like Run(), but get the output as a string
func RunGetOut(command string, args ...string) string {
	var err error
	var out []byte
	if out, err = runImpl(true, command, args...); err != nil {
		glog.Fatalf("%s: %s", command, err)
	}
	return strings.TrimSpace(string(out))
}

