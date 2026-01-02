package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cesc1802/janus/internal/migrator"
)

var historyLimit int

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show migration history",
	Long:  `Display list of migrations with their applied status.`,
	RunE:  runHistory,
}

func init() {
	historyCmd.Flags().IntVar(&historyLimit, "limit", 10, "Number of migrations to show")
	rootCmd.AddCommand(historyCmd)
}

func runHistory(cmd *cobra.Command, args []string) error {
	mg, err := migrator.New(envName)
	if err != nil {
		return err
	}
	defer func() { _ = mg.Close() }()

	status, err := mg.Status()
	if err != nil {
		return err
	}

	migrations := mg.GetMigrationList(status.Version)

	fmt.Printf("Migration History (env: %s)\n", envName)
	fmt.Println("----------------------------------------")

	if len(migrations) == 0 {
		fmt.Println("  No migrations found")
		return nil
	}

	// Show up to limit migrations
	shown := 0
	for _, m := range migrations {
		if shown >= historyLimit {
			break
		}

		marker := "[ ]"
		if m.Applied {
			marker = "[x]"
		}
		fmt.Printf("  %s %06d - %s\n", marker, m.Version, m.Name)
		shown++
	}

	if len(migrations) > historyLimit {
		fmt.Printf("\n  ... and %d more (use --limit to show more)\n", len(migrations)-historyLimit)
	}

	return nil
}
