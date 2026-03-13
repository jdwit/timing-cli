package cmd

import (
	"fmt"
	"os"

	"github.com/jdwit/timing-cli/internal/api"
	"github.com/jdwit/timing-cli/internal/config"
	"github.com/jdwit/timing-cli/internal/output"
	"github.com/spf13/cobra"
)

var jsonFlag bool

var rootCmd = &cobra.Command{
	Use:   "timing",
	Short: "CLI for the Timing macOS time tracking app",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		output.JSONOutput = jsonFlag
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "output in JSON format")

	rootCmd.AddCommand(projectsCmd)
	rootCmd.AddCommand(entriesCmd)
	rootCmd.AddCommand(timerCmd)
	rootCmd.AddCommand(activitiesCmd)
	rootCmd.AddCommand(reportCmd)
	rootCmd.AddCommand(completionCmd)
}

var newClient = func() *api.Client {
	apiKey := config.GetAPIKey()
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: TIMING_API_KEY environment variable not set")
		os.Exit(1)
	}
	return api.NewClient(apiKey, config.GetTimezone())
}
