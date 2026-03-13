package cmd

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/jdwit/timing-cli/internal/output"
	"github.com/spf13/cobra"
)

var activitiesCmd = &cobra.Command{
	Use:   "activities",
	Short: "Show activity hierarchy",
	Long:  "Returns a hierarchical view of activities grouped by project and application.",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		start, _ := cmd.Flags().GetString("start")
		end, _ := cmd.Flags().GetString("end")
		blockSize, _ := cmd.Flags().GetString("block-size")
		minDuration, _ := cmd.Flags().GetInt("min-duration")
		groupByProject, _ := cmd.Flags().GetBool("group-by-project")
		maxDepth, _ := cmd.Flags().GetInt("max-depth")
		maxLines, _ := cmd.Flags().GetInt("max-lines")
		includeMobile, _ := cmd.Flags().GetBool("include-mobile")
		projectIDs, _ := cmd.Flags().GetStringSlice("project-ids")
		includeSubprojects, _ := cmd.Flags().GetBool("include-subprojects")

		params := url.Values{}
		if start != "" {
			params.Set("start_date", start)
		}
		if end != "" {
			params.Set("end_date", end)
		}
		if blockSize != "" {
			params.Set("block_size", blockSize)
		}
		if cmd.Flags().Changed("min-duration") {
			params.Set("minimum_duration_seconds", strconv.Itoa(minDuration))
		}
		if cmd.Flags().Changed("group-by-project") {
			params.Set("group_by_project", strconv.FormatBool(groupByProject))
		}
		if cmd.Flags().Changed("max-depth") {
			params.Set("max_depth", strconv.Itoa(maxDepth))
		}
		if cmd.Flags().Changed("max-lines") {
			params.Set("max_lines", strconv.Itoa(maxLines))
		}
		if includeMobile {
			params.Set("include_mobile_devices", "true")
		}
		for _, id := range projectIDs {
			params.Add("project_ids[]", id)
		}
		if cmd.Flags().Changed("include-subprojects") {
			params.Set("include_subprojects", strconv.FormatBool(includeSubprojects))
		}

		text, err := c.GetText("/activity-hierarchy", params)
		if err != nil {
			return err
		}

		if output.JSONOutput {
			output.PrintJSON(map[string]string{"activity_hierarchy": text})
			return nil
		}

		fmt.Print(text)
		return nil
	},
}

func init() {
	activitiesCmd.Flags().String("start", "", "start date (required, e.g. 2024-01-01)")
	activitiesCmd.Flags().String("end", "", "end date (required, e.g. 2024-01-31)")
	activitiesCmd.Flags().String("block-size", "total", "granularity: total, month, week, day, hour, 15min, 5min")
	activitiesCmd.Flags().Int("min-duration", 60, "minimum activity duration in seconds")
	activitiesCmd.Flags().Bool("group-by-project", true, "group activities by project")
	activitiesCmd.Flags().Int("max-depth", 0, "maximum nesting depth (0 = unlimited)")
	activitiesCmd.Flags().Int("max-lines", 100, "maximum number of leaf activities (1-1000)")
	activitiesCmd.Flags().Bool("include-mobile", false, "include iPhone/iPad data")
	activitiesCmd.Flags().StringSlice("project-ids", nil, "filter to specific project IDs (use 0 for unassigned)")
	activitiesCmd.Flags().Bool("include-subprojects", true, "include child projects when filtering")
}
