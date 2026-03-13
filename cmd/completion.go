package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion scripts for timing.

To load completions:

Bash:
  $ source <(timing completion bash)
  # Or for permanent setup:
  $ timing completion bash > /etc/bash_completion.d/timing

Zsh:
  $ timing completion zsh > "${fpath[1]}/_timing"
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		}
		return nil
	},
}
