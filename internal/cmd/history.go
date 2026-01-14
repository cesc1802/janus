package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/cesc1802/janus/internal/migrator"
)

var historyLimit int

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show migration history",
	Long:  `Display list of migrations with their applied status and history.`,
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

	// Fetch rich history
	historyLogs, err := mg.History().GetHistory(0)
	if err != nil {
		// If history table doesn't exist or fails, just warn but continue?
		// Or maybe just show empty history.
		fmt.Fprintf(os.Stderr, "Warning: could not fetch detailed history: %v\n", err)
	}

	// Map version -> latest successful UP entry
	historyMap := make(map[uint]migrator.HistoryEntry)
	for _, entry := range historyLogs {
		if entry.Action == "up" && entry.Success {
			// Since logs are ordered DESC (latest first), the first one we see is the latest
			if _, exists := historyMap[entry.Version]; !exists {
				historyMap[entry.Version] = entry
			}
		}
	}

	fmt.Printf("Migration History (env: %s)\n", envName)
	fmt.Println("----------------------------------------")

	if len(migrations) == 0 {
		fmt.Println("  No migrations found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "STATUS\tVERSION\tNAME\tAPPLIED AT\tDURATION")

	// Show up to limit migrations
	shown := 0
	for _, m := range migrations {
		if shown >= historyLimit {
			break
		}

		statusMarker := "[ ]"
		appliedAt := "-"
		duration := "-"

		if m.Applied {
			statusMarker = "[x]"
			if entry, ok := historyMap[m.Version]; ok {
				appliedAt = entry.StartTime.Format("2006-01-02 15:04:05")
				duration = entry.Duration.String()
			}
		}

		fmt.Fprintf(w, "  %s\t%06d\t%s\t%s\t%s\n", statusMarker, m.Version, m.Name, appliedAt, duration)
		shown++
	}
	w.Flush()

	if len(migrations) > historyLimit {
		fmt.Printf("\n  ... and %d more (use --limit to show more)\n", len(migrations)-historyLimit)
	}

	return nil
}
