package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new migration file",
	Long: `Create a new migration file with up/down template.

Examples:
  migrate-tool create create_users_table
  migrate-tool create add_email_to_users`,
	Args: cobra.ExactArgs(1),
	RunE: runCreate,
}

var createSeq bool

func init() {
	createCmd.Flags().BoolVar(&createSeq, "seq", true, "Use sequential versioning (vs timestamp)")
	rootCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Sanitize name
	name = sanitizeName(name)
	if name == "" {
		return fmt.Errorf("invalid migration name")
	}

	// Limit name length
	if len(name) > 100 {
		return fmt.Errorf("migration name too long (max 100 chars)")
	}

	// Get migrations path from config
	migrationsPath := viper.GetString("defaults.migrations_path")
	if migrationsPath == "" {
		migrationsPath = "./migrations"
	}

	// Security: resolve to absolute path for validation
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("resolve migrations path: %w", err)
	}

	// Ensure dir exists
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return fmt.Errorf("create migrations dir: %w", err)
	}

	// Generate version
	var ver string
	if createSeq {
		ver = getNextSequentialVersion(absPath)
	} else {
		ver = fmt.Sprintf("%d", time.Now().Unix())
	}

	// Create file
	filename := fmt.Sprintf("%s_%s.sql", ver, name)
	fpath := filepath.Join(absPath, filename)

	// Security: verify file stays within migrations directory
	absFpath, err := filepath.Abs(fpath)
	if err != nil {
		return fmt.Errorf("resolve file path: %w", err)
	}
	if !strings.HasPrefix(absFpath, absPath+string(filepath.Separator)) && absFpath != absPath {
		return fmt.Errorf("invalid file path: path traversal detected")
	}

	content := migrationTemplate(name)
	// Security: owner read/write only (0600) for migration files
	if err := os.WriteFile(fpath, []byte(content), 0600); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	fmt.Printf("Created: %s\n", fpath)
	return nil
}

func sanitizeName(name string) string {
	// Replace spaces and special chars with underscore
	re := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	name = re.ReplaceAllString(name, "_")
	name = strings.Trim(name, "_")
	return strings.ToLower(name)
}

func getNextSequentialVersion(dir string) string {
	entries, _ := os.ReadDir(dir)

	maxVersion := 0
	pattern := regexp.MustCompile(`^(\d+)_`)

	for _, entry := range entries {
		if matches := pattern.FindStringSubmatch(entry.Name()); matches != nil {
			if v, _ := strconv.Atoi(matches[1]); v > maxVersion {
				maxVersion = v
			}
		}
	}

	return fmt.Sprintf("%06d", maxVersion+1)
}

func migrationTemplate(name string) string {
	return fmt.Sprintf(`-- Migration: %s
-- Created: %s

-- +migrate UP
-- TODO: Add your UP migration SQL here


-- +migrate DOWN
-- TODO: Add your DOWN migration SQL here

`, name, time.Now().Format("2006-01-02 15:04:05"))
}
