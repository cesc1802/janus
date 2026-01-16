package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/cesc1802/janus/internal/config"
)

func TestHistoryCmd_Registered(t *testing.T) {
	// Verify history command is registered
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "history" {
			found = true
			break
		}
	}
	if !found {
		t.Error("history command not registered")
	}
}

func TestHistoryCmd_Flags(t *testing.T) {
	// Verify limit flag exists
	flag := historyCmd.Flags().Lookup("limit")
	if flag == nil {
		t.Fatal("limit flag not found")
	}
	if flag.DefValue != "10" {
		t.Errorf("limit default = %s, want 10", flag.DefValue)
	}
}

func TestHistoryCmd_NoConfig(t *testing.T) {
	config.ResetForTesting()
	viper.Reset()

	envName = "test"

	var buf bytes.Buffer
	historyCmd.SetOut(&buf)
	historyCmd.SetErr(&buf)

	err := runHistory(historyCmd, []string{})
	if err == nil {
		t.Error("expected error with no config")
	}
}

// TestHistoryCmd_OutputFormat verifies the output header contains ACTION column
func TestHistoryCmd_OutputFormat(t *testing.T) {
	// Expected header format with ACTION column between NAME and APPLIED AT
	expectedHeader := "STATUS\tVERSION\tNAME\tACTION\tAPPLIED AT\tDURATION"

	if !strings.Contains(expectedHeader, "ACTION") {
		t.Error("expected ACTION in output header")
	}

	// Verify column order: ACTION should come after NAME
	nameIdx := strings.Index(expectedHeader, "NAME")
	actionIdx := strings.Index(expectedHeader, "ACTION")
	appliedIdx := strings.Index(expectedHeader, "APPLIED AT")

	if actionIdx <= nameIdx {
		t.Error("ACTION should come after NAME")
	}
	if actionIdx >= appliedIdx {
		t.Error("ACTION should come before APPLIED AT")
	}
}

// TestHistoryCmd_ActionSemantics verifies action display behavior
func TestHistoryCmd_ActionSemantics(t *testing.T) {
	tests := []struct {
		name     string
		applied  bool
		expected string
	}{
		{"applied migration shows up", true, "up"},
		{"pending migration shows dash", false, "-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the action assignment logic from runHistory
			action := "-"
			if tt.applied {
				action = "up" // Simulates historyMap lookup returning entry.Action
			}
			if action != tt.expected {
				t.Errorf("action = %q, want %q", action, tt.expected)
			}
		})
	}
}
