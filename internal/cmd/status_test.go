package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/viper"

	"github.com/cesc1802/janus/internal/config"
)

func TestStatusCmd_Registered(t *testing.T) {
	// Verify status command is registered
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "status" {
			found = true
			break
		}
	}
	if !found {
		t.Error("status command not registered")
	}
}

func TestStatusCmd_NoConfig(t *testing.T) {
	config.ResetForTesting()
	viper.Reset()

	envName = "test"

	var buf bytes.Buffer
	statusCmd.SetOut(&buf)
	statusCmd.SetErr(&buf)

	err := runStatus(statusCmd, []string{})
	if err == nil {
		t.Error("expected error with no config")
	}
}
