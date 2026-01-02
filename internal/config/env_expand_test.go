package config

import (
	"os"
	"testing"
)

func TestExpandEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		envKey   string
		envVal   string
		expected string
	}{
		{
			name:     "expands single var",
			input:    "${TEST_DB_URL}",
			envKey:   "TEST_DB_URL",
			envVal:   "postgres://localhost:5432/testdb",
			expected: "postgres://localhost:5432/testdb",
		},
		{
			name:     "expands var with suffix",
			input:    "${TEST_DB_URL}?sslmode=disable",
			envKey:   "TEST_DB_URL",
			envVal:   "postgres://localhost:5432/testdb",
			expected: "postgres://localhost:5432/testdb?sslmode=disable",
		},
		{
			name:     "keeps unexpanded if not set",
			input:    "${NONEXISTENT_VAR}",
			envKey:   "",
			envVal:   "",
			expected: "${NONEXISTENT_VAR}",
		},
		{
			name:     "no vars to expand",
			input:    "postgres://localhost:5432/db",
			envKey:   "",
			envVal:   "",
			expected: "postgres://localhost:5432/db",
		},
		{
			name:     "expands multiple vars",
			input:    "${TEST_HOST}:${TEST_PORT}",
			envKey:   "TEST_HOST",
			envVal:   "localhost",
			expected: "localhost:${TEST_PORT}",
		},
		{
			name:     "empty string",
			input:    "",
			envKey:   "",
			envVal:   "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envKey != "" {
				_ = os.Setenv(tt.envKey, tt.envVal)
				defer func(key string) { _ = os.Unsetenv(key) }(tt.envKey)
			}

			result := ExpandEnvVars(tt.input)
			if result != tt.expected {
				t.Errorf("ExpandEnvVars(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExpandEnvVars_MultipleVars(t *testing.T) {
	_ = os.Setenv("TEST_MULTI_HOST", "myhost")
	_ = os.Setenv("TEST_MULTI_PORT", "5432")
	defer func() {
		_ = os.Unsetenv("TEST_MULTI_HOST")
		_ = os.Unsetenv("TEST_MULTI_PORT")
	}()

	input := "postgres://${TEST_MULTI_HOST}:${TEST_MULTI_PORT}/db"
	expected := "postgres://myhost:5432/db"

	result := ExpandEnvVars(input)
	if result != expected {
		t.Errorf("ExpandEnvVars(%q) = %q, want %q", input, result, expected)
	}
}
