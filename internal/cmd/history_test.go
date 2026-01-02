package cmd

import (
	"bytes"
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
