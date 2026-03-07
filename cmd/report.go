package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/jdwit/timing-cli/internal/api"
	"github.com/jdwit/timing-cli/internal/output"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a time report",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		start, _ := cmd.Flags().GetString("start")
		end, _ := cmd.Flags().GetString("end")
		projects, _ := cmd.Flags().GetStringSlice("project")
		includeChildren, _ := cmd.Flags().GetBool("include-children")
		search, _ := cmd.Flags().GetString("search")
		columns, _ := cmd.Flags().GetStringSlice("columns")
		groupLevel, _ := cmd.Flags().GetInt("project-grouping-level")
		timespanGrouping, _ := cmd.Flags().GetString("timespan-grouping")
		billing, _ := cmd.Flags().GetStringSlice("billing-status")
		sort, _ := cmd.Flags().GetStringSlice("sort")
		includeAppUsage, _ := cmd.Flags().GetBool("include-app-usage")
		includeTeamMembers, _ := cmd.Flags().GetBool("include-team-members")
		teamMembers, _ := cmd.Flags().GetStringSlice("team-members")

		params := url.Values{}
		params.Set("include_project_data", "true")
		if start != "" {
			params.Set("start_date_min", start)
		}
		if end != "" {
			params.Set("start_date_max", end)
		}
		for _, p := range projects {
			params.Add("projects[]", p)
		}
		if includeChildren {
			params.Set("include_child_projects", "true")
		}
		if search != "" {
			params.Set("search_query", search)
		}
		if len(columns) > 0 {
			for _, col := range columns {
				params.Add("columns[]", col)
			}
		}
		if cmd.Flags().Changed("project-grouping-level") {
			params.Set("project_grouping_level", strconv.Itoa(groupLevel))
		}
		if timespanGrouping != "" {
			params.Set("timespan_grouping_mode", timespanGrouping)
		}
		for _, b := range billing {
			params.Add("billing_status[]", b)
		}
		for _, s := range sort {
			params.Add("sort[]", s)
		}
		if includeAppUsage {
			params.Set("include_app_usage", "true")
		}
		if includeTeamMembers {
			params.Set("include_team_members", "true")
		}
		for _, m := range teamMembers {
			params.Add("team_members[]", m)
		}

		body, err := c.Get("/report", params)
		if err != nil {
			return err
		}

		var resp struct {
			Data []api.ReportRow `json:"data"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if output.JSONOutput {
			output.PrintJSON(resp.Data)
			return nil
		}

		headers := []string{"DURATION", "PROJECT", "TITLE"}
		var rows [][]string
		totalDuration := 0.0
		for _, r := range resp.Data {
			projectName := ""
			if r.Project != nil {
				projectName = strings.Join(r.Project.TitleChain, " > ")
				if projectName == "" {
					projectName = r.Project.Self
				}
			}
			totalDuration += r.Duration
			rows = append(rows, []string{
				output.FormatDuration(r.Duration),
				output.Truncate(projectName, 40),
				output.Truncate(r.Title, 40),
			})
		}
		output.PrintTable(headers, rows)
		fmt.Printf("\nTotal: %s\n", output.FormatDuration(totalDuration))
		return nil
	},
}

func init() {
	reportCmd.Flags().String("start", "", "start date (ISO 8601)")
	reportCmd.Flags().String("end", "", "end date (ISO 8601)")
	reportCmd.Flags().StringSlice("project", nil, "filter by project references")
	reportCmd.Flags().Bool("include-children", false, "include child project entries")
	reportCmd.Flags().String("search", "", "search time entry titles/notes")
	reportCmd.Flags().StringSlice("columns", nil, "report columns: project, title, notes, timespan, user, billing_status")
	reportCmd.Flags().Int("project-grouping-level", -1, "aggregate projects at level (-1 = no grouping)")
	reportCmd.Flags().String("timespan-grouping", "", "grouping: exact, day, week, month, year")
	reportCmd.Flags().StringSlice("billing-status", nil, "filter by billing status")
	reportCmd.Flags().StringSlice("sort", nil, "sort columns (prefix - for descending)")
	reportCmd.Flags().Bool("include-app-usage", false, "include app usage data")
	reportCmd.Flags().Bool("include-team-members", false, "include team members' data")
	reportCmd.Flags().StringSlice("team-members", nil, "restrict to specific team members")
}
