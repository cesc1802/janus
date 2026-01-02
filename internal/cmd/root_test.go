package cmd

import (
	"testing"
)

func TestSetVersionInfo(t *testing.T) {
	SetVersionInfo("1.0.0", "abc123", "2026-01-01")

	if version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", version)
	}
	if commit != "abc123" {
		t.Errorf("expected commit 'abc123', got '%s'", commit)
	}
	if date != "2026-01-01" {
		t.Errorf("expected date '2026-01-01', got '%s'", date)
	}
}

func TestExecute(t *testing.T) {
	// Test that Execute doesn't panic
	err := Execute()
	if err != nil {
		t.Errorf("Execute() returned error: %v", err)
	}
}

func TestRootCmdExists(t *testing.T) {
	if rootCmd == nil {
		t.Error("rootCmd is nil")
	}
	if rootCmd.Use != "janus" {
		t.Errorf("expected Use 'janus', got '%s'", rootCmd.Use)
	}
}
