package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestRunVersion(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set version info
	SetVersionInfo("1.0.0", "abc123", "2026-01-01")

	// Run version command
	runVersion(nil, nil)

	// Restore stdout
	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Verify output contains expected info
	if !strings.Contains(output, "janus 1.0.0") {
		t.Errorf("output should contain version: %s", output)
	}
	if !strings.Contains(output, "commit: abc123") {
		t.Errorf("output should contain commit: %s", output)
	}
	if !strings.Contains(output, "built:  2026-01-01") {
		t.Errorf("output should contain build date: %s", output)
	}
	if !strings.Contains(output, "go:") {
		t.Errorf("output should contain go version: %s", output)
	}
	if !strings.Contains(output, "os:") {
		t.Errorf("output should contain os info: %s", output)
	}
}

func TestRunVersion_DevDefaults(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set empty version info (simulating dev build)
	SetVersionInfo("", "", "")

	// Run version command
	runVersion(nil, nil)

	// Restore stdout
	_ = w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Verify default values are used
	if !strings.Contains(output, "janus dev") {
		t.Errorf("output should show 'dev' for empty version: %s", output)
	}
	if !strings.Contains(output, "commit: unknown") {
		t.Errorf("output should show 'unknown' for empty commit: %s", output)
	}
	if !strings.Contains(output, "built:  unknown") {
		t.Errorf("output should show 'unknown' for empty date: %s", output)
	}
}
