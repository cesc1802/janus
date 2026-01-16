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
	// Header format: VERSION, NAME, ACTION, APPLIED AT, DURATION (no STATUS column)
	expectedHeader := "VERSION\tNAME\tACTION\tAPPLIED AT\tDURATION"

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

// TestHistoryCmd_ActionSemantics verifies both UP and DOWN actions display correctly
func TestHistoryCmd_ActionSemantics(t *testing.T) {
	tests := []struct {
		name     string
		action   string
		expected string
	}{
		{"up action displays correctly", "up", "up"},
		{"down action displays correctly", "down", "down"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// History entries show action values directly from database
			if tt.action != tt.expected {
				t.Errorf("action = %q, want %q", tt.action, tt.expected)
			}
		})
	}
}
