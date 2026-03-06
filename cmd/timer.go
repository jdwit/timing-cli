package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jdwit/timing-cli/internal/api"
	"github.com/jdwit/timing-cli/internal/output"
	"github.com/spf13/cobra"
)

var timerCmd = &cobra.Command{
	Use:   "timer",
	Short: "Manage the running timer",
}

var timerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a timer",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		project, _ := cmd.Flags().GetString("project")
		title, _ := cmd.Flags().GetString("title")
		notes, _ := cmd.Flags().GetString("notes")
		startDate, _ := cmd.Flags().GetString("start")
		replace, _ := cmd.Flags().GetBool("replace-existing")
		billing, _ := cmd.Flags().GetString("billing-status")

		payload := map[string]any{}
		if project != "" {
			payload["project"] = project
		}
		if title != "" {
			payload["title"] = title
		}
		if notes != "" {
			payload["notes"] = notes
		}
		if startDate != "" {
			payload["start_date"] = startDate
		}
		if replace {
			payload["replace_existing"] = true
		}
		if billing != "" {
			payload["billing_status"] = billing
		}

		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		body, err := c.Post("/time-entries/start", bytes.NewReader(data))
		if err != nil {
			return err
		}

		var resp struct {
			Data    api.TimeEntry `json:"data"`
			Message string        `json:"message"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if output.JSONOutput {
			output.PrintJSON(resp)
			return nil
		}

		if resp.Message != "" {
			fmt.Println(resp.Message)
		} else {
			title := resp.Data.Title
			if title == "" {
				title = "(untitled)"
			}
			fmt.Printf("Timer started: %s (%s)\n", title, resp.Data.Self)
		}
		return nil
	},
}

var timerStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the running timer",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		body, err := c.Put("/time-entries/stop", nil)
		if err != nil {
			return err
		}

		var resp struct {
			Data api.TimeEntry `json:"data"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if output.JSONOutput {
			output.PrintJSON(resp.Data)
			return nil
		}

		fmt.Printf("Timer stopped: %s (%s)\n", resp.Data.Title, output.FormatDuration(resp.Data.Duration))
		return nil
	},
}

var timerStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the running timer",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		body, err := c.GetWithRedirect("/time-entries/running", nil)
		if err != nil {
			return err
		}

		if body == nil {
			if output.JSONOutput {
				output.PrintJSON(map[string]any{"running": false})
			} else {
				fmt.Println("No timer running.")
			}
			return nil
		}

		var resp struct {
			Data api.TimeEntry `json:"data"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if output.JSONOutput {
			output.PrintJSON(resp.Data)
			return nil
		}

		e := resp.Data
		title := e.Title
		if title == "" {
			title = "(untitled)"
		}
		fmt.Printf("Timer running: %s\n", title)
		fmt.Printf("Started:       %s\n", e.StartDate)
		if e.Project != nil {
			name := strings.Join(e.Project.TitleChain, " > ")
			if name == "" {
				name = e.Project.Self
			}
			fmt.Printf("Project:       %s\n", name)
		}
		return nil
	},
}

func init() {
	timerStartCmd.Flags().String("project", "", "project reference (e.g. /projects/1)")
	timerStartCmd.Flags().String("title", "", "timer title")
	timerStartCmd.Flags().String("notes", "", "timer notes")
	timerStartCmd.Flags().String("start", "", "custom start date/time (ISO 8601)")
	timerStartCmd.Flags().Bool("replace-existing", false, "replace existing overlapping entries")
	timerStartCmd.Flags().String("billing-status", "", "billing status")

	timerCmd.AddCommand(timerStartCmd)
	timerCmd.AddCommand(timerStopCmd)
	timerCmd.AddCommand(timerStatusCmd)
}
