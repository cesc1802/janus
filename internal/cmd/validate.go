package cmd

import (
	"fmt"
	"os"

	"github.com/cesc1802/migrate-tool/internal/config"
	"github.com/cesc1802/migrate-tool/internal/source/singlefile"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration and migration files",
	Long: `Validate configuration file and migration files for syntax errors.

Examples:
  migrate-tool validate
  migrate-tool validate --env=prod`,
	RunE: runValidate,
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	var errors []string
	var warnings []string

	// 1. Validate config
	fmt.Println("Validating configuration...")
	cfg, err := config.Load()
	if err != nil {
		errors = append(errors, fmt.Sprintf("Config: %v", err))
	} else {
		fmt.Printf("  Found %d environment(s)\n", len(cfg.Environments))
	}

	// 2. Validate migrations for specified env or all envs if none specified
	var envs []string
	if envName != "" && cfg != nil {
		// User specified an env, validate only that one
		envs = append(envs, envName)
	} else if cfg != nil {
		// No env specified, validate all
		for name := range cfg.Environments {
			envs = append(envs, name)
		}
	}

	for _, env := range envs {
		fmt.Printf("\nValidating migrations for '%s'...\n", env)

		envCfg, err := config.GetEnv(env)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Env %s: %v", env, err))
			continue
		}

		// Check path exists
		if _, err := os.Stat(envCfg.MigrationsPath); os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("Env %s: migrations path not found: %s", env, envCfg.MigrationsPath))
			continue
		}

		// Try to load migrations
		driver, err := singlefile.NewWithPath(envCfg.MigrationsPath)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Env %s: %v", env, err))
			continue
		}

		// Count and validate migrations
		count := 0
		emptyUp := 0
		emptyDown := 0

		v, err := driver.First()
		for err == nil {
			count++

			// Check for empty up/down
			upReader, _, upErr := driver.ReadUp(v)
			if upErr != nil {
				emptyUp++
			} else {
				upReader.Close()
			}

			downReader, _, downErr := driver.ReadDown(v)
			if downErr != nil {
				emptyDown++
			} else {
				downReader.Close()
			}

			v, err = driver.Next(v)
		}

		fmt.Printf("  Found %d migration(s)\n", count)

		if emptyUp > 0 {
			warnings = append(warnings, fmt.Sprintf("Env %s: %d migration(s) with empty UP section", env, emptyUp))
		}
		if emptyDown > 0 {
			warnings = append(warnings, fmt.Sprintf("Env %s: %d migration(s) with empty DOWN section", env, emptyDown))
		}
	}

	// Print summary
	fmt.Println("\n─────────────────────────────")
	if len(errors) > 0 {
		fmt.Println("ERRORS:")
		for _, e := range errors {
			fmt.Printf("  ✗ %s\n", e)
		}
	}
	if len(warnings) > 0 {
		fmt.Println("WARNINGS:")
		for _, w := range warnings {
			fmt.Printf("  ! %s\n", w)
		}
	}

	if len(errors) == 0 && len(warnings) == 0 {
		fmt.Println("✓ All validations passed")
		return nil
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed with %d error(s)", len(errors))
	}

	return nil
}
