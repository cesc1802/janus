package cmd

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/cesc1802/migrate-tool/internal/config"
	"github.com/spf13/viper"
)

func TestGotoCmd_Registered(t *testing.T) {
	// Verify goto command is registered
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "goto <version>" {
			found = true
			break
		}
	}
	if !found {
		t.Error("goto command not registered")
	}
}

func TestGotoCmd_RequiresArg(t *testing.T) {
	// goto requires exactly 1 argument
	if gotoCmd.Args == nil {
		t.Error("goto command should have Args validator")
	}
}

func TestGotoCmd_VersionParsing(t *testing.T) {
	tests := []struct {
		input string
		valid bool
		value uint64
	}{
		{"0", true, 0},
		{"1", true, 1},
		{"10", true, 10},
		{"100", true, 100},
		{"abc", false, 0},
		{"-1", false, 0}, // negative not valid for goto (uses ParseUint)
		{"1.5", false, 0},
	}

	for _, tc := range tests {
		v, err := strconv.ParseUint(tc.input, 10, 64)
		if tc.valid && err != nil {
			t.Errorf("%s: expected valid, got error: %v", tc.input, err)
		}
		if !tc.valid && err == nil {
			t.Errorf("%s: expected invalid, got valid", tc.input)
		}
		if tc.valid && v != tc.value {
			t.Errorf("%s: got %d, want %d", tc.input, v, tc.value)
		}
	}
}

func TestGotoCmd_NoConfig(t *testing.T) {
	config.ResetForTesting()
	viper.Reset()

	envName = "test"

	var buf bytes.Buffer
	gotoCmd.SetOut(&buf)
	gotoCmd.SetErr(&buf)

	err := runGoto(gotoCmd, []string{"5"})
	if err == nil {
		t.Error("expected error with no config")
	}
}

func TestGotoCmd_InvalidVersion(t *testing.T) {
	err := runGoto(gotoCmd, []string{"abc"})
	if err == nil {
		t.Error("expected error with invalid version")
	}
}

func TestGotoCmd_NegativeVersion(t *testing.T) {
	err := runGoto(gotoCmd, []string{"-1"})
	if err == nil {
		t.Error("expected error with negative version")
	}
}
