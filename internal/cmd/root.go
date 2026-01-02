package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile               string
	envName               string
	autoApprove           bool
	version, commit, date string
	configLoaded          bool
)

var rootCmd = &cobra.Command{
	Use:   "janus",
	Short: "Database migration CLI tool",
	Long:  `Cross-platform database migration tool with single-file up/down support.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// SetVersionInfo sets version information for the CLI
func SetVersionInfo(v, c, d string) {
	version, commit, date = v, c, d
}

// GetEnvName returns the current environment name
func GetEnvName() string {
	return envName
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ./janus.yaml)")
	rootCmd.PersistentFlags().StringVar(&envName, "env", "dev", "environment name")
	rootCmd.PersistentFlags().BoolVar(&autoApprove, "auto-approve", false, "skip confirmation prompts (for CI/CD)")
}

// AutoApprove returns whether confirmation prompts should be skipped
func AutoApprove() bool {
	return autoApprove
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("janus")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	}
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
			os.Exit(1)
		}
		// Config file not found is OK for some commands (version, help)
	} else {
		configLoaded = true
	}
}

// IsConfigLoaded returns whether a config file was found and loaded
func IsConfigLoaded() bool {
	return configLoaded
}
