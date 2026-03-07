package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jdwit/timing-cli/internal/config"
	"github.com/jdwit/timing-cli/internal/output"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a config value",
	Long:  "Set a configuration value. Valid keys: api-key, timezone",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Set(args[0], args[1]); err != nil {
			return err
		}
		if !output.JSONOutput {
			fmt.Printf("Set %s\n", args[0])
		} else {
			output.PrintJSON(map[string]string{"key": args[0], "value": args[1]})
		}
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a config value",
	Long:  "Get a configuration value. Valid keys: api-key, timezone",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		val, err := config.Get(args[0])
		if err != nil {
			return err
		}
		if output.JSONOutput {
			output.PrintJSON(map[string]string{"key": args[0], "value": val})
		} else {
			fmt.Println(val)
		}
		return nil
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactive configuration setup",
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		cfg, _ := config.Load()

		fmt.Print("API key: ")
		apiKey, _ := reader.ReadString('\n')
		apiKey = strings.TrimSpace(apiKey)
		if apiKey != "" {
			cfg.APIKey = apiKey
		}

		fmt.Printf("Timezone [%s]: ", cfg.Timezone)
		tz, _ := reader.ReadString('\n')
		tz = strings.TrimSpace(tz)
		if tz != "" {
			cfg.Timezone = tz
		}

		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Printf("Config saved to %s\n", config.File())
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configInitCmd)
}
