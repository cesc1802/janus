package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/viper"

	"github.com/cesc1802/janus/internal/config"
)

func TestDownCmd_Registered(t *testing.T) {
	// Verify down command is registered
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "down" {
			found = true
			break
		}
	}
	if !found {
		t.Error("down command not registered")
	}
}

func TestDownCmd_Flags(t *testing.T) {
	// Verify steps flag exists
	flag := downCmd.Flags().Lookup("steps")
	if flag == nil {
		t.Fatal("steps flag not found")
	}
	if flag.DefValue != "1" {
		t.Errorf("steps default = %s, want 1", flag.DefValue)
	}
}

func TestDownCmd_NoConfig(t *testing.T) {
	config.ResetForTesting()
	viper.Reset()

	envName = "test"

	var buf bytes.Buffer
	downCmd.SetOut(&buf)
	downCmd.SetErr(&buf)

	err := runDown(downCmd, []string{})
	if err == nil {
		t.Error("expected error with no config")
	}
}
