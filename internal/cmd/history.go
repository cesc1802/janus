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

	// Fetch history (ordered by start_time DESC - latest first)
	historyLogs, err := mg.History().GetHistory(0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not fetch history: %v\n", err)
		historyLogs = []migrator.HistoryEntry{}
	}

	fmt.Printf("Migration History (env: %s)\n", envName)
	fmt.Println("----------------------------------------")

	if len(historyLogs) == 0 {
		fmt.Println("  No migration history found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "VERSION\tNAME\tACTION\tAPPLIED AT\tDURATION")

	// Show up to limit history entries (chronologically, latest first)
	shown := 0
	for _, entry := range historyLogs {
		if shown >= historyLimit {
			break
		}
		_, _ = fmt.Fprintf(w, "%06d\t%s\t%s\t%s\t%s\n",
			entry.Version,
			entry.Name,
			entry.Action,
			entry.StartTime.Format("2006-01-02 15:04:05"),
			entry.Duration.String(),
		)
		shown++
	}
	_ = w.Flush()

	if len(historyLogs) > historyLimit {
		fmt.Printf("\n  ... and %d more (use --limit to show more)\n", len(historyLogs)-historyLimit)
	}

	return nil
}
