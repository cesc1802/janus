package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cesc1802/migrate-tool/internal/migrator"
	"github.com/cesc1802/migrate-tool/internal/ui"
)

var forceCmd = &cobra.Command{
	Use:   "force <version>",
	Short: "Force set migration version (for dirty state recovery)",
	Long: `Force set the migration version without running any migrations.

USE WITH CAUTION: This is intended for recovering from dirty state
after a failed migration. It does NOT run any migration code.

Examples:
  migrate-tool force 5 --env=dev     # Set version to 5
  migrate-tool force 0 --env=dev     # Reset to initial state
  migrate-tool force -1 --env=dev    # Clear version (NilVersion)`,
	Args: cobra.ExactArgs(1),
	RunE: runForce,
}

func init() {
	rootCmd.AddCommand(forceCmd)
}

func runForce(cmd *cobra.Command, args []string) error {
	version, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid version: %w", err)
	}

	mg, err := migrator.New(envName)
	if err != nil {
		return err
	}
	defer func() { _ = mg.Close() }()

	// Get current status for context
	status, err := mg.Status()
	if err != nil {
		return fmt.Errorf("get status: %w", err)
	}

	fmt.Println("WARNING: Force Version Change")
	fmt.Printf("Environment: %s\n", envName)
	fmt.Printf("Current version: %d (dirty: %v)\n", status.Version, status.Dirty)
	fmt.Printf("New version: %d\n\n", version)
	fmt.Println("This will NOT run any migrations.")
	fmt.Println("Use this only to recover from dirty state.")
	fmt.Println()

	// Confirmation logic
	if !AutoApprove() {
		details := fmt.Sprintf("Force setting version from %d to %d\nThis does NOT run migrations", status.Version, version)

		if mg.RequiresConfirmation() {
			confirmed, err := ui.ConfirmProduction(envName)
			if err != nil {
				return err
			}
			if !confirmed {
				ui.Warning("Cancelled")
				return nil
			}
		} else {
			confirmed, err := ui.ConfirmDangerous("force version", details)
			if err != nil {
				return err
			}
			if !confirmed {
				ui.Warning("Cancelled")
				return nil
			}
		}
	}

	if err := mg.Force(version); err != nil {
		return fmt.Errorf("force failed: %w", err)
	}

	ui.Success(fmt.Sprintf("Version forced to %d", version))
	return nil
}
