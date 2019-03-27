package cmd

import (
	"io/ioutil"
	"os"
	"testing"
)

func mustCompareOSImageURL(t *testing.T, refA, refB string) bool {
	m, err := compareOSImageURL(refA, refB)
	if err != nil {
		t.Fatalf("%v", err)
	}
	return m
}

func TestCompareOSImageURL(t *testing.T) {
	refA := "registry.example.com/foo/bar@sha256:0743a3cc3bcf3b4aabb814500c2739f84cb085ff4e7ec7996aef7977c4c19c7f"
	refB := "registry.example.com/foo/baz@sha256:0743a3cc3bcf3b4aabb814500c2739f84cb085ff4e7ec7996aef7977c4c19c7f"
	refC := "registry.example.com/foo/bar@sha256:2a76681fd15bfc06fa4aa0ff6913ba17527e075417fc92ea29f6bcc2afca24ff"
	if !mustCompareOSImageURL(t, refA, refA) {
		t.Fatalf("Expected refA ident")
	}
	if !mustCompareOSImageURL(t, refA, refB) {
		t.Fatalf("Expected refA = refB")
	}
	if mustCompareOSImageURL(t, refA, refC) {
		t.Fatalf("Expected refA != refC")
	}
	m, err := compareOSImageURL(refA, "registry.example.com/foo/bar")
	if m || err == nil {
		t.Fatalf("Expected err")
	}
}

// writeTestFile writes out a file to use in the test
func writeTestFile(content []byte) (filePath string, err error) {
	tmpfile, err := ioutil.TempFile("", "testFile")
	if err != nil {
		return "", err
	}
	filePath = tmpfile.Name()
	if _, err := tmpfile.Write(content); err != nil {
		return filePath, err
	}
	if err := tmpfile.Close(); err != nil {
		return filePath, err
	}
	return filePath, nil
}

func TestParseTuningFile(t *testing.T) {
	cmdLineFileMock, err := writeTestFile([]byte(
		"BOOT_IMAGE=/a/vmlinuz.x86_64 resume=/dev/mapper/swap rhgb quiet root=/a/b/c/root ostree=/ostree/boot.0/a/0"))
	defer os.Remove(cmdLineFileMock)

	// Test with addition/deletion and verify white list
	testFilePath, err := writeTestFile([]byte("ADD nosmt\nADD aaaa\nDELETE nosmt\nDELETE nope"))
	defer os.Remove(testFilePath)
	if err != nil {
		t.Fatalf("unable to write test file %s: %s", testFilePath, err)
	}
	add, delete, err := parseTuningFile(testFilePath, cmdLineFileMock)
	if err != nil {
		t.Fatalf(`Expected no error, got %s`, err)
	}
	if len(add) != 1 {
		t.Fatalf("Expected 1 addition, got %v", len(add))
	}

	if len(delete) != 0 {
		t.Fatalf("Expected 0 deletion, got %v", len(delete))
	}

	deleteCmdLineFileMockWith, err := writeTestFile([]byte(
		"BOOT_IMAGE=/a/vmlinuz.x86_64 nosmt resume=/dev/mapper/swap rhgb quiet root=/a/b/c/root ostree=/ostree/boot.0/a/0"))
	defer os.Remove(deleteCmdLineFileMockWith)

	// Test with addition/deletion and verify white list
	testFilePath, err = writeTestFile([]byte("ADD nosmt\nADD aaaa\nDELETE nosmt\nDELETE nope"))
	defer os.Remove(testFilePath)
	if err != nil {
		t.Fatalf("unable to write test file %s: %s", testFilePath, err)
	}
	add, delete, err = parseTuningFile(testFilePath, deleteCmdLineFileMockWith)
	if err != nil {
		t.Fatalf(`Expected no error, got %s`, err)
	}
	if len(add) != 0 {
		t.Fatalf("Expected 1 addition, got %v", len(add))
	}

	if len(delete) != 1 {
		t.Fatalf("Expected 1 deletion, got %v", len(delete))
	}

	// Test with no changes
	testFilePath, err = writeTestFile([]byte(""))
	defer os.Remove(testFilePath)
	if err != nil {
		t.Fatalf("unable to write test file %s: %s", testFilePath, err)
	}
	add, delete, err = parseTuningFile(testFilePath, cmdLineFileMock)
	if err != nil {
		t.Fatalf(`Expected no error, got %s`, err)
	}

	if len(add) != 0 {
		t.Fatalf("Expected 0 addition, got %v", len(add))
	}

	if len(delete) != 0 {
		t.Fatalf("Expected 0 deletion, got %v", len(add))
	}
}

func TestIsArgInUse(t *testing.T) {
	testFilePath, err := writeTestFile([]byte(
		"BOOT_IMAGE=/a/vmlinuz.x86_64 resume=/dev/mapper/swap rhgb quiet root=/a/b/c/root ostree=/ostree/boot.0/a/0"))
	if err != nil {
		t.Fatalf("unable to write test file %s: %s", testFilePath, err)
	}
	defer os.Remove(testFilePath)

	// Should be present
	available, err := isArgInUse("quiet", testFilePath)
	if err != nil {
		t.Fatalf(`Expected no error, got %s`, err)
	}
	if available != true {
		t.Fatalf("Expected true, got false")
	}

	// Should not be present
	available, err = isArgInUse("idonotexist", testFilePath)
	if err != nil {
		t.Fatalf(`Expected no error, got %s`, err)
	}
	if available != false {
		t.Fatalf("Expected false, got true")
	}
}
