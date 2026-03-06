package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/jdwit/timing-cli/internal/api"
	"github.com/jdwit/timing-cli/internal/output"
	"github.com/spf13/cobra"
)

var entriesCmd = &cobra.Command{
	Use:     "entries",
	Aliases: []string{"entry", "e"},
	Short:   "Manage time entries",
}

var entriesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List time entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		start, _ := cmd.Flags().GetString("start")
		end, _ := cmd.Flags().GetString("end")
		project, _ := cmd.Flags().GetString("project")
		search, _ := cmd.Flags().GetString("search")
		isRunning, _ := cmd.Flags().GetString("is-running")
		includeChildren, _ := cmd.Flags().GetBool("include-children")
		teamMembers, _ := cmd.Flags().GetBool("include-team-members")
		billing, _ := cmd.Flags().GetStringSlice("billing-status")
		pageLimit, _ := cmd.Flags().GetInt("page-limit")

		params := url.Values{}
		params.Set("include_project_data", "true")
		if start != "" {
			params.Set("start_date_min", start)
		}
		if end != "" {
			params.Set("start_date_max", end)
		}
		if project != "" {
			params.Add("projects[]", project)
		}
		if search != "" {
			params.Set("search_query", search)
		}
		if isRunning != "" {
			params.Set("is_running", isRunning)
		}
		if includeChildren {
			params.Set("include_child_projects", "true")
		}
		if teamMembers {
			params.Set("include_team_members", "true")
		}
		for _, b := range billing {
			params.Add("billing_status[]", b)
		}

		items, err := c.GetPaginated("/time-entries", params, pageLimit)
		if err != nil {
			return err
		}

		var entries []api.TimeEntry
		for _, item := range items {
			var e api.TimeEntry
			if err := json.Unmarshal(item, &e); err != nil {
				return fmt.Errorf("parsing entry: %w", err)
			}
			entries = append(entries, e)
		}

		if output.JSONOutput {
			output.PrintJSON(entries)
			return nil
		}

		headers := []string{"ID", "START", "DURATION", "PROJECT", "TITLE"}
		var rows [][]string
		for _, e := range entries {
			projectName := ""
			if e.Project != nil {
				projectName = strings.Join(e.Project.TitleChain, " > ")
				if projectName == "" {
					projectName = e.Project.Self
				}
			}
			dur := output.FormatDuration(e.Duration)
			if e.IsRunning {
				dur = "running"
			}
			rows = append(rows, []string{
				e.Self,
				formatDate(e.StartDate),
				dur,
				output.Truncate(projectName, 30),
				output.Truncate(e.Title, 40),
			})
		}
		output.PrintTable(headers, rows)
		return nil
	},
}

func formatDate(iso string) string {
	if len(iso) >= 16 {
		return iso[:16]
	}
	return iso
}

var entriesGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Show time entry details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()
		id := normalizeRef(args[0], "time-entries")

		params := url.Values{}
		params.Set("include_project_data", "true")

		body, err := c.Get("/time-entries/"+id, params)
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

		e := resp.Data
		fmt.Printf("ID:       %s\n", e.Self)
		fmt.Printf("Start:    %s\n", e.StartDate)
		fmt.Printf("End:      %s\n", e.EndDate)
		fmt.Printf("Duration: %s\n", output.FormatDuration(e.Duration))
		if e.Project != nil {
			name := strings.Join(e.Project.TitleChain, " > ")
			if name == "" {
				name = e.Project.Self
			}
			fmt.Printf("Project:  %s\n", name)
		}
		if e.Title != "" {
			fmt.Printf("Title:    %s\n", e.Title)
		}
		if e.Notes != "" {
			fmt.Printf("Notes:    %s\n", e.Notes)
		}
		fmt.Printf("Running:  %v\n", e.IsRunning)
		fmt.Printf("Billing:  %s\n", e.BillingStatus)
		return nil
	},
}

var entriesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a time entry",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		start, _ := cmd.Flags().GetString("start")
		end, _ := cmd.Flags().GetString("end")
		project, _ := cmd.Flags().GetString("project")
		title, _ := cmd.Flags().GetString("title")
		notes, _ := cmd.Flags().GetString("notes")
		replace, _ := cmd.Flags().GetBool("replace-existing")
		billing, _ := cmd.Flags().GetString("billing-status")

		payload := map[string]any{
			"start_date": start,
			"end_date":   end,
		}
		if project != "" {
			payload["project"] = project
		}
		if title != "" {
			payload["title"] = title
		}
		if notes != "" {
			payload["notes"] = notes
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

		body, err := c.Post("/time-entries", bytes.NewReader(data))
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

		fmt.Printf("Created time entry: %s (%s)\n", resp.Data.Self, output.FormatDuration(resp.Data.Duration))
		return nil
	},
}

var entriesUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a time entry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()
		id := normalizeRef(args[0], "time-entries")

		payload := map[string]any{}
		if cmd.Flags().Changed("start") {
			v, _ := cmd.Flags().GetString("start")
			payload["start_date"] = v
		}
		if cmd.Flags().Changed("end") {
			v, _ := cmd.Flags().GetString("end")
			payload["end_date"] = v
		}
		if cmd.Flags().Changed("project") {
			v, _ := cmd.Flags().GetString("project")
			payload["project"] = v
		}
		if cmd.Flags().Changed("title") {
			v, _ := cmd.Flags().GetString("title")
			payload["title"] = v
		}
		if cmd.Flags().Changed("notes") {
			v, _ := cmd.Flags().GetString("notes")
			payload["notes"] = v
		}
		if cmd.Flags().Changed("replace-existing") {
			v, _ := cmd.Flags().GetBool("replace-existing")
			payload["replace_existing"] = v
		}
		if cmd.Flags().Changed("billing-status") {
			v, _ := cmd.Flags().GetString("billing-status")
			payload["billing_status"] = v
		}

		if len(payload) == 0 {
			return fmt.Errorf("no fields to update; use flags like --title, --notes, --project")
		}

		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		body, err := c.Put("/time-entries/"+id, bytes.NewReader(data))
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

		fmt.Printf("Updated time entry: %s\n", resp.Data.Self)
		return nil
	},
}

var entriesDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a time entry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()
		id := normalizeRef(args[0], "time-entries")

		if err := c.Delete("/time-entries/" + id); err != nil {
			return err
		}

		if output.JSONOutput {
			output.PrintJSON(map[string]string{"status": "deleted", "id": id})
		} else {
			fmt.Printf("Deleted time entry %s\n", id)
		}
		return nil
	},
}

var entriesBatchUpdateCmd = &cobra.Command{
	Use:   "batch-update",
	Short: "Bulk update time entries",
	Long:  "Update multiple time entries at once. Provide entry IDs with --entries and fields to update.",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		entryIDs, _ := cmd.Flags().GetStringSlice("entries")
		if len(entryIDs) == 0 {
			return fmt.Errorf("--entries is required")
		}

		dataPayload := map[string]any{}
		if cmd.Flags().Changed("project") {
			v, _ := cmd.Flags().GetString("project")
			dataPayload["project"] = v
		}
		if cmd.Flags().Changed("title") {
			v, _ := cmd.Flags().GetString("title")
			dataPayload["title"] = v
		}
		if cmd.Flags().Changed("notes") {
			v, _ := cmd.Flags().GetString("notes")
			dataPayload["notes"] = v
		}
		if cmd.Flags().Changed("billing-status") {
			v, _ := cmd.Flags().GetString("billing-status")
			dataPayload["billing_status"] = v
		}

		allowOthers, _ := cmd.Flags().GetBool("allow-editing-other-users")

		// Normalize entry IDs to bare IDs
		var ids []string
		for _, id := range entryIDs {
			ids = append(ids, normalizeRef(id, "time-entries"))
		}

		payload := map[string]any{
			"time_entries":              ids,
			"data":                      dataPayload,
			"allow_editing_other_users": allowOthers,
		}

		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		body, err := c.Patch("/time-entries/batch-update", bytes.NewReader(data))
		if err != nil {
			return err
		}

		var resp struct {
			Data []api.TimeEntry `json:"data"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if output.JSONOutput {
			output.PrintJSON(resp.Data)
			return nil
		}

		fmt.Printf("Updated %d time entries\n", len(resp.Data))
		return nil
	},
}

func init() {
	entriesListCmd.Flags().String("start", "", "start date (ISO 8601)")
	entriesListCmd.Flags().String("end", "", "end date (ISO 8601)")
	entriesListCmd.Flags().String("project", "", "filter by project reference")
	entriesListCmd.Flags().String("search", "", "search titles and notes")
	entriesListCmd.Flags().String("is-running", "", "filter by running state (true/false)")
	entriesListCmd.Flags().Bool("include-children", false, "include child project entries")
	entriesListCmd.Flags().Bool("include-team-members", false, "include team members' entries")
	entriesListCmd.Flags().StringSlice("billing-status", nil, "filter by billing status")
	entriesListCmd.Flags().Int("page-limit", 0, "maximum number of pages to fetch (0 = all)")

	entriesCreateCmd.Flags().String("start", "", "start date/time (ISO 8601, required)")
	entriesCreateCmd.MarkFlagRequired("start")
	entriesCreateCmd.Flags().String("end", "", "end date/time (ISO 8601, required)")
	entriesCreateCmd.MarkFlagRequired("end")
	entriesCreateCmd.Flags().String("project", "", "project reference (e.g. /projects/1)")
	entriesCreateCmd.Flags().String("title", "", "entry title")
	entriesCreateCmd.Flags().String("notes", "", "entry notes")
	entriesCreateCmd.Flags().Bool("replace-existing", false, "replace existing overlapping entries")
	entriesCreateCmd.Flags().String("billing-status", "", "billing status")

	entriesGetCmd.Flags().String("other-user-id", "", "other user ID for team entries")

	entriesUpdateCmd.Flags().String("start", "", "start date/time (ISO 8601)")
	entriesUpdateCmd.Flags().String("end", "", "end date/time (ISO 8601)")
	entriesUpdateCmd.Flags().String("project", "", "project reference")
	entriesUpdateCmd.Flags().String("title", "", "entry title")
	entriesUpdateCmd.Flags().String("notes", "", "entry notes")
	entriesUpdateCmd.Flags().Bool("replace-existing", false, "replace existing overlapping entries")
	entriesUpdateCmd.Flags().String("billing-status", "", "billing status")

	entriesBatchUpdateCmd.Flags().StringSlice("entries", nil, "time entry IDs to update (required)")
	entriesBatchUpdateCmd.Flags().String("project", "", "project reference")
	entriesBatchUpdateCmd.Flags().String("title", "", "entry title")
	entriesBatchUpdateCmd.Flags().String("notes", "", "entry notes")
	entriesBatchUpdateCmd.Flags().String("billing-status", "", "billing status")
	entriesBatchUpdateCmd.Flags().Bool("allow-editing-other-users", false, "allow editing other users' entries")

	entriesCmd.AddCommand(entriesListCmd)
	entriesCmd.AddCommand(entriesGetCmd)
	entriesCmd.AddCommand(entriesCreateCmd)
	entriesCmd.AddCommand(entriesUpdateCmd)
	entriesCmd.AddCommand(entriesDeleteCmd)
	entriesCmd.AddCommand(entriesBatchUpdateCmd)
}
