package singlefile

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	// filenamePattern matches migration files: {version}_{name}.sql
	filenamePattern = regexp.MustCompile(`^(\d+)_(.+)\.sql$`)
	upMarker        = "-- +migrate UP"
	downMarker      = "-- +migrate DOWN"
)

// Migration represents a parsed migration file
type Migration struct {
	Version uint
	Name    string
	Up      string
	Down    string
}

// parseMigrationFile reads and parses a single migration file
func parseMigrationFile(path string) (Migration, error) {
	filename := filepath.Base(path)
	matches := filenamePattern.FindStringSubmatch(filename)
	if matches == nil {
		return Migration{}, fmt.Errorf("invalid migration filename: %s (expected format: {version}_{name}.sql)", filename)
	}

	version, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return Migration{}, fmt.Errorf("invalid version in filename %s: %w", filename, err)
	}
	name := matches[2]

	content, err := os.ReadFile(path)
	if err != nil {
		return Migration{}, fmt.Errorf("read migration file %s: %w", filename, err)
	}

	up, down := parseContent(string(content))

	return Migration{
		Version: uint(version),
		Name:    name,
		Up:      up,
		Down:    down,
	}, nil
}

// parseContent extracts UP and DOWN sections from migration content
func parseContent(content string) (up, down string) {
	var upLines, downLines []string
	var currentSection string

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(trimmed, upMarker):
			currentSection = "up"
			continue
		case strings.HasPrefix(trimmed, downMarker):
			currentSection = "down"
			continue
		}

		switch currentSection {
		case "up":
			upLines = append(upLines, line)
		case "down":
			downLines = append(downLines, line)
		}
	}

	return strings.TrimSpace(strings.Join(upLines, "\n")),
		strings.TrimSpace(strings.Join(downLines, "\n"))
}

// validateFilename checks if a filename matches migration pattern
func validateFilename(filename string) bool {
	return filenamePattern.MatchString(filename)
}
