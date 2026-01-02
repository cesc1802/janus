package ui

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestUseColor(t *testing.T) {
	// In test environment, stdout is typically not a TTY
	// UseColor should return false in non-TTY environments
	result := UseColor()
	// Just verify it doesn't panic and returns a boolean
	if result && os.Getenv("NO_COLOR") != "" {
		t.Error("UseColor should return false when NO_COLOR is set")
	}
}

func TestSuccess(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Success("test message")

	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = oldStdout

	output := string(out)
	if !strings.Contains(output, "test message") {
		t.Errorf("Success should contain message, got: %s", output)
	}
	if !strings.Contains(output, "OK") {
		t.Errorf("Success should contain OK, got: %s", output)
	}
}

func TestWarning(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Warning("warning test")

	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = oldStdout

	output := string(out)
	if !strings.Contains(output, "warning test") {
		t.Errorf("Warning should contain message, got: %s", output)
	}
	if !strings.Contains(output, "!") {
		t.Errorf("Warning should contain !, got: %s", output)
	}
}

func TestError(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	Error("error test")

	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stderr = oldStderr

	output := string(out)
	if !strings.Contains(output, "error test") {
		t.Errorf("Error should contain message, got: %s", output)
	}
	if !strings.Contains(output, "ERROR") {
		t.Errorf("Error should contain ERROR, got: %s", output)
	}
}

func TestInfo(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Info("info test")

	_ = w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = oldStdout

	output := string(out)
	if !strings.Contains(output, "info test") {
		t.Errorf("Info should contain message, got: %s", output)
	}
	if !strings.Contains(output, "*") {
		t.Errorf("Info should contain *, got: %s", output)
	}
}

func TestColorConstants(t *testing.T) {
	// Verify color constants are properly defined
	tests := []struct {
		name  string
		color string
	}{
		{"ColorReset", ColorReset},
		{"ColorRed", ColorRed},
		{"ColorGreen", ColorGreen},
		{"ColorYellow", ColorYellow},
		{"ColorBlue", ColorBlue},
		{"ColorBold", ColorBold},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.color == "" {
				t.Errorf("%s should not be empty", tt.name)
			}
			if !strings.HasPrefix(tt.color, "\033[") {
				t.Errorf("%s should be ANSI escape code, got: %s", tt.name, tt.color)
			}
		})
	}
}

func TestNoColorEnv(t *testing.T) {
	// Test NO_COLOR environment variable behavior
	originalNoColor := os.Getenv("NO_COLOR")
	defer func() { _ = os.Setenv("NO_COLOR", originalNoColor) }()

	_ = os.Setenv("NO_COLOR", "1")
	if UseColor() {
		t.Error("UseColor should return false when NO_COLOR is set")
	}
}
