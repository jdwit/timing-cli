package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/jdwit/timing-cli/internal/api"
	"github.com/jdwit/timing-cli/internal/output"
	"github.com/spf13/cobra"
)

var teamsCmd = &cobra.Command{
	Use:   "teams",
	Short: "Manage teams",
}

var teamsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all teams",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		body, err := c.Get("/teams", nil)
		if err != nil {
			return err
		}

		var resp struct {
			Data []api.Team `json:"data"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if output.JSONOutput {
			output.PrintJSON(resp.Data)
			return nil
		}

		headers := []string{"ID", "NAME"}
		var rows [][]string
		for _, t := range resp.Data {
			rows = append(rows, []string{t.ID, t.Name})
		}
		output.PrintTable(headers, rows)
		return nil
	},
}

var teamsMembersCmd = &cobra.Command{
	Use:   "members <team_id>",
	Short: "List team members",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()
		id := normalizeRef(args[0], "teams")

		body, err := c.Get("/teams/"+id+"/members", nil)
		if err != nil {
			return err
		}

		var resp struct {
			Data []api.TeamMember `json:"data"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if output.JSONOutput {
			output.PrintJSON(resp.Data)
			return nil
		}

		headers := []string{"ID", "NAME", "EMAIL"}
		var rows [][]string
		for _, m := range resp.Data {
			rows = append(rows, []string{m.Self, m.Name, m.Email})
		}
		output.PrintTable(headers, rows)
		return nil
	},
}

func init() {
	teamsCmd.AddCommand(teamsListCmd)
	teamsCmd.AddCommand(teamsMembersCmd)
}
