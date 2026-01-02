package cmd

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/spf13/cobra"

	"github.com/cesc1802/migrate-tool/internal/migrator"
	"github.com/cesc1802/migrate-tool/internal/ui"
)

var downSteps int

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback migrations",
	Long:  `Rollback the last migration or specified number of steps.`,
	RunE:  runDown,
}

func init() {
	downCmd.Flags().IntVar(&downSteps, "steps", 1, "Number of migrations to rollback")
	rootCmd.AddCommand(downCmd)
}

func runDown(cmd *cobra.Command, args []string) error {
	mg, err := migrator.New(envName)
	if err != nil {
		return err
	}
	defer func() { _ = mg.Close() }()

	// Get status before rollback
	status, err := mg.Status()
	if err != nil {
		return fmt.Errorf("get status: %w", err)
	}

	if status.Applied == 0 {
		ui.Info("No migrations to rollback")
		return nil
	}

	// Show what will happen
	fmt.Printf("Environment: %s\n", envName)
	fmt.Printf("Current version: %d\n", status.Version)
	fmt.Printf("Will rollback: %d migration(s)\n", downSteps)
	fmt.Println()

	// Confirmation logic
	if !AutoApprove() {
		details := fmt.Sprintf("Rolling back %d migration(s) in %s", downSteps, envName)

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
			confirmed, err := ui.ConfirmDangerous("rollback", details)
			if err != nil {
				return err
			}
			if !confirmed {
				ui.Warning("Cancelled")
				return nil
			}
		}
	}

	if err := mg.Down(downSteps); err != nil {
		if err == migrate.ErrNoChange {
			ui.Info("No migrations to rollback")
			return nil
		}
		return fmt.Errorf("rollback failed: %w", err)
	}

	// Get status after rollback
	newStatus, err := mg.Status()
	if err != nil {
		return fmt.Errorf("get status: %w", err)
	}

	rolledBack := status.Applied - newStatus.Applied
	ui.Success(fmt.Sprintf("Rolled back %d migration(s)", rolledBack))
	if newStatus.Version > 0 {
		fmt.Printf("Current version: %d\n", newStatus.Version)
	} else {
		fmt.Println("Current version: none (clean slate)")
	}

	return nil
}
