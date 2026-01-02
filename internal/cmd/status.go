package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cesc1802/migrate-tool/internal/migrator"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	Long:  `Display current migration version, dirty state, and pending count.`,
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	mg, err := migrator.New(envName)
	if err != nil {
		return err
	}
	defer func() { _ = mg.Close() }()

	status, err := mg.Status()
	if err != nil {
		return err
	}

	fmt.Printf("Environment: %s\n", envName)
	if status.Version > 0 {
		fmt.Printf("Current Version: %d\n", status.Version)
	} else {
		fmt.Println("Current Version: none (no migrations applied)")
	}
	fmt.Printf("Dirty: %v\n", status.Dirty)
	fmt.Printf("Applied: %d / %d\n", status.Applied, status.Total)
	fmt.Printf("Pending: %d\n", status.Pending)

	if status.Dirty {
		fmt.Println("\nWARNING: Database is in dirty state.")
		fmt.Println("This usually means a migration failed mid-execution.")
		fmt.Printf("Fix with: migrate-tool force %d --env=%s\n", status.Version, envName)
	}

	return nil
}
