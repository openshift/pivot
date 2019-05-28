package cmd

import (
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
