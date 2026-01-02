package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display janus version, commit hash, build date, and Go runtime info.`,
	Run:   runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command, args []string) {
	v := version
	if v == "" {
		v = "dev"
	}
	c := commit
	if c == "" {
		c = "unknown"
	}
	d := date
	if d == "" {
		d = "unknown"
	}

	fmt.Printf("janus %s\n", v)
	fmt.Printf("  commit: %s\n", c)
	fmt.Printf("  built:  %s\n", d)
	fmt.Printf("  go:     %s\n", runtime.Version())
	fmt.Printf("  os:     %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
