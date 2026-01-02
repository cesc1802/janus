package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/viper"

	"github.com/cesc1802/janus/internal/config"
)

func TestUpCmd_Registered(t *testing.T) {
	// Verify up command is registered
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "up" {
			found = true
			break
		}
	}
	if !found {
		t.Error("up command not registered")
	}
}

func TestUpCmd_Flags(t *testing.T) {
	// Verify steps flag exists
	flag := upCmd.Flags().Lookup("steps")
	if flag == nil {
		t.Fatal("steps flag not found")
	}
	if flag.DefValue != "0" {
		t.Errorf("steps default = %s, want 0", flag.DefValue)
	}
}

func TestUpCmd_NoConfig(t *testing.T) {
	config.ResetForTesting()
	viper.Reset()

	// Don't set any config
	envName = "test"

	var buf bytes.Buffer
	upCmd.SetOut(&buf)
	upCmd.SetErr(&buf)

	err := runUp(upCmd, []string{})
	if err == nil {
		t.Error("expected error with no config")
	}
}
