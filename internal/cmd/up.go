package cmd

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/spf13/cobra"

	"github.com/cesc1802/janus/internal/migrator"
	"github.com/cesc1802/janus/internal/ui"
)

var upSteps int

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply pending migrations",
	Long:  `Apply all pending migrations or specified number of steps.`,
	RunE:  runUp,
}

func init() {
	upCmd.Flags().IntVar(&upSteps, "steps", 0, "Number of migrations to apply (0 = all)")
	rootCmd.AddCommand(upCmd)
}

func runUp(cmd *cobra.Command, args []string) error {
	mg, err := migrator.New(envName)
	if err != nil {
		return err
	}
	defer func() { _ = mg.Close() }()

	// Get status before migration
	status, err := mg.Status()
	if err != nil {
		return fmt.Errorf("get status: %w", err)
	}

	if status.Pending == 0 {
		ui.Info("No pending migrations")
		return nil
	}

	// Show what will happen
	fmt.Printf("Environment: %s\n", envName)
	fmt.Printf("Pending migrations: %d\n", status.Pending)
	if upSteps > 0 {
		fmt.Printf("Will apply: %d migration(s)\n", upSteps)
	} else {
		fmt.Printf("Will apply: all %d migration(s)\n", status.Pending)
	}
	fmt.Println()

	// Confirmation logic
	if !AutoApprove() {
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
			confirmed, err := ui.Confirm("Apply migrations?", false)
			if err != nil {
				return err
			}
			if !confirmed {
				ui.Warning("Cancelled")
				return nil
			}
		}
	}

	if err := mg.Up(upSteps); err != nil {
		if err == migrate.ErrNoChange {
			ui.Info("No migrations to apply")
			return nil
		}
		return fmt.Errorf("migration failed: %w", err)
	}

	// Get status after migration
	newStatus, err := mg.Status()
	if err != nil {
		return fmt.Errorf("get status: %w", err)
	}

	applied := newStatus.Applied - status.Applied
	ui.Success(fmt.Sprintf("Applied %d migration(s)", applied))
	fmt.Printf("Current version: %d\n", newStatus.Version)

	return nil
}
