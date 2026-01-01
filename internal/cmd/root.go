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
	version, commit, date string
	configLoaded          bool
)

var rootCmd = &cobra.Command{
	Use:   "migrate-tool",
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ./migrate-tool.yaml)")
	rootCmd.PersistentFlags().StringVar(&envName, "env", "dev", "environment name")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("migrate-tool")
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
