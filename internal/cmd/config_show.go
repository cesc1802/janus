package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cesc1802/janus/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current configuration with sensitive values masked.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if !IsConfigLoaded() {
			return fmt.Errorf("no config file found (looking for janus.yaml)")
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		fmt.Printf("Config file: %s\n\n", "janus.yaml")
		fmt.Println("Environments:")
		for name, env := range cfg.Environments {
			fmt.Printf("  %s:\n", name)
			fmt.Printf("    database_url: %s\n", maskDatabaseURL(env.DatabaseURL))
			fmt.Printf("    migrations_path: %s\n", env.MigrationsPath)
			fmt.Printf("    require_confirmation: %v\n", env.RequireConfirmation)
		}

		if cfg.Defaults.MigrationsPath != "" || cfg.Defaults.RequireConfirmation {
			fmt.Println("\nDefaults:")
			if cfg.Defaults.MigrationsPath != "" {
				fmt.Printf("  migrations_path: %s\n", cfg.Defaults.MigrationsPath)
			}
			fmt.Printf("  require_confirmation: %v\n", cfg.Defaults.RequireConfirmation)
		}

		return nil
	},
}

// maskDatabaseURL masks password in database URLs for security
// postgres://user:password@host:port/db -> postgres://user:***@host:port/db
func maskDatabaseURL(url string) string {
	schemeEnd := strings.Index(url, "://")
	if schemeEnd == -1 {
		return url
	}

	// Find the @ that separates credentials from host (last @ before path/query)
	afterScheme := url[schemeEnd+3:]
	atIdx := strings.LastIndex(afterScheme, "@")
	if atIdx == -1 {
		return url
	}

	// Find user:password part between :// and @
	userPassPart := afterScheme[:atIdx]
	colonIdx := strings.Index(userPassPart, ":")
	if colonIdx == -1 {
		return url // No password
	}

	user := userPassPart[:colonIdx]
	return url[:schemeEnd+3] + user + ":***@" + afterScheme[atIdx+1:]
}

func init() {
	configCmd.AddCommand(configShowCmd)
	rootCmd.AddCommand(configCmd)
}
