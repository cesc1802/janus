package cmd

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/spf13/viper"

	"github.com/cesc1802/janus/internal/config"
)

func TestForceCmd_Registered(t *testing.T) {
	// Verify force command is registered
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "force <version>" {
			found = true
			break
		}
	}
	if !found {
		t.Error("force command not registered")
	}
}

func TestForceCmd_RequiresArg(t *testing.T) {
	// force requires exactly 1 argument
	if forceCmd.Args == nil {
		t.Error("force command should have Args validator")
	}
}

func TestForceCmd_VersionParsing(t *testing.T) {
	tests := []struct {
		input   string
		valid   bool
		version int
	}{
		{"5", true, 5},
		{"0", true, 0},
		{"-1", true, -1},
		{"100", true, 100},
		{"abc", false, 0},
		{"1.5", false, 0},
		{"", false, 0},
	}

	for _, tc := range tests {
		v, err := strconv.Atoi(tc.input)
		if tc.valid && err != nil {
			t.Errorf("%s: expected valid, got error: %v", tc.input, err)
		}
		if !tc.valid && err == nil {
			t.Errorf("%s: expected invalid, got valid", tc.input)
		}
		if tc.valid && v != tc.version {
			t.Errorf("%s: got %d, want %d", tc.input, v, tc.version)
		}
	}
}

func TestForceCmd_NoConfig(t *testing.T) {
	config.ResetForTesting()
	viper.Reset()

	envName = "test"

	var buf bytes.Buffer
	forceCmd.SetOut(&buf)
	forceCmd.SetErr(&buf)

	err := runForce(forceCmd, []string{"5"})
	if err == nil {
		t.Error("expected error with no config")
	}
}

func TestForceCmd_InvalidVersion(t *testing.T) {
	err := runForce(forceCmd, []string{"abc"})
	if err == nil {
		t.Error("expected error with invalid version")
	}
}
