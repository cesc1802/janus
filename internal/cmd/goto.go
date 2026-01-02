package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	"github.com/spf13/cobra"

	"github.com/cesc1802/migrate-tool/internal/migrator"
	"github.com/cesc1802/migrate-tool/internal/ui"
)

var gotoCmd = &cobra.Command{
	Use:   "goto <version>",
	Short: "Migrate to a specific version",
	Long: `Migrate up or down to reach the specified version.

If target version > current: applies UP migrations
If target version < current: applies DOWN migrations

Examples:
  migrate-tool goto 10 --env=dev    # Migrate to version 10
  migrate-tool goto 0 --env=dev     # Rollback all migrations`,
	Args: cobra.ExactArgs(1),
	RunE: runGoto,
}

func init() {
	rootCmd.AddCommand(gotoCmd)
}

func runGoto(cmd *cobra.Command, args []string) error {
	targetVersion, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid version: %w", err)
	}

	mg, err := migrator.New(envName)
	if err != nil {
		return err
	}
	defer func() { _ = mg.Close() }()

	// Get current status
	status, err := mg.Status()
	if err != nil {
		return fmt.Errorf("get status: %w", err)
	}

	// Check dirty state - cannot migrate if database is dirty
	if status.Dirty {
		ui.Warning("Database is in dirty state.")
		fmt.Println("Use 'migrate-tool force <version>' to fix the dirty state first.")
		return fmt.Errorf("cannot migrate: database in dirty state at version %d", status.Version)
	}

	// Determine direction and count
	var direction string
	var stepsCount int
	target := uint(targetVersion)

	if target > status.Version {
		direction = "UP"
		stepsCount = countMigrationsBetween(mg, status.Version, target)
	} else if target < status.Version {
		direction = "DOWN"
		stepsCount = countMigrationsBetween(mg, target, status.Version)
	} else {
		ui.Info(fmt.Sprintf("Already at version %d", targetVersion))
		return nil
	}

	fmt.Println("Migration Target")
	fmt.Printf("Environment: %s\n", envName)
	fmt.Printf("Current version: %d\n", status.Version)
	fmt.Printf("Target version: %d\n", targetVersion)
	fmt.Printf("Direction: %s (%d migration(s))\n\n", direction, stepsCount)

	// Confirmation logic
	if !AutoApprove() {
		details := fmt.Sprintf("Migrating from %d to %d (%s, %d migrations)", status.Version, targetVersion, direction, stepsCount)

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
			confirmed, err := ui.ConfirmDangerous("goto", details)
			if err != nil {
				return err
			}
			if !confirmed {
				ui.Warning("Cancelled")
				return nil
			}
		}
	}

	if err := mg.Goto(target); err != nil {
		if err == migrate.ErrNoChange {
			ui.Info("No migrations to apply")
			return nil
		}
		return fmt.Errorf("goto failed: %w", err)
	}

	ui.Success(fmt.Sprintf("Migrated to version %d", targetVersion))
	return nil
}

// countMigrationsBetween counts migrations between from and to versions (exclusive from, inclusive to).
// Note: This differs from migrator.countMigrations which categorizes applied/pending.
// This function counts range for display purposes only. On error, returns 0 (display-only impact).
func countMigrationsBetween(mg *migrator.Migrator, from, to uint) int {
	count := 0
	src := mg.Source()

	v, err := src.First()
	for err == nil {
		if v > from && v <= to {
			count++
		}
		v, err = src.Next(v)
	}

	// Handle ErrNotExist gracefully
	if err != nil && err != os.ErrNotExist {
		return 0
	}

	return count
}
