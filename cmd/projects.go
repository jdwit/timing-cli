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

var projectsCmd = &cobra.Command{
	Use:     "projects",
	Aliases: []string{"project", "p"},
	Short:   "Manage projects",
}

var projectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		tree, _ := cmd.Flags().GetBool("tree")
		title, _ := cmd.Flags().GetString("title")
		hideArchived, _ := cmd.Flags().GetBool("hide-archived")
		teamID, _ := cmd.Flags().GetString("team-id")

		params := url.Values{}
		if title != "" {
			params.Set("title", title)
		}
		if hideArchived {
			params.Set("hide_archived", "1")
		}
		if teamID != "" {
			params.Set("team_id", teamID)
		}

		var path string
		if tree {
			path = "/projects/hierarchy"
		} else {
			path = "/projects"
		}

		body, err := c.Get(path, params)
		if err != nil {
			return err
		}

		var resp struct {
			Data []api.Project `json:"data"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if output.JSONOutput {
			output.PrintJSON(resp.Data)
			return nil
		}

		if tree {
			printProjectTree(resp.Data, 0)
		} else {
			headers := []string{"ID", "TITLE", "COLOR", "ARCHIVED", "PARENT"}
			var rows [][]string
			for _, p := range resp.Data {
				parent := ""
				if p.Parent != nil {
					parent = p.Parent.Self
				}
				archived := ""
				if p.IsArchived {
					archived = "yes"
				}
				rows = append(rows, []string{
					p.Self,
					strings.Join(p.TitleChain, " > "),
					p.Color,
					archived,
					parent,
				})
			}
			output.PrintTable(headers, rows)
		}
		return nil
	},
}

func printProjectTree(projects []api.Project, depth int) {
	for _, p := range projects {
		indent := strings.Repeat("  ", depth)
		archived := ""
		if p.IsArchived {
			archived = " [archived]"
		}
		fmt.Printf("%s%s (%s)%s\n", indent, p.Title, p.Self, archived)
		if len(p.Children) > 0 {
			printProjectTree(p.Children, depth+1)
		}
	}
}

var projectsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Show project details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()
		id := normalizeRef(args[0], "projects")

		body, err := c.Get("/projects/"+id, nil)
		if err != nil {
			return err
		}

		var resp struct {
			Data api.Project `json:"data"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if output.JSONOutput {
			output.PrintJSON(resp.Data)
			return nil
		}

		p := resp.Data
		fmt.Printf("ID:       %s\n", p.Self)
		fmt.Printf("Title:    %s\n", strings.Join(p.TitleChain, " > "))
		fmt.Printf("Color:    %s\n", p.Color)
		fmt.Printf("Archived: %v\n", p.IsArchived)
		if p.Parent != nil {
			fmt.Printf("Parent:   %s\n", p.Parent.Self)
		}
		if p.Notes != nil && *p.Notes != "" {
			fmt.Printf("Notes:    %s\n", *p.Notes)
		}
		fmt.Printf("Billing:  %s\n", p.DefaultBillingStatus)
		if p.ProductivityScore != nil {
			fmt.Printf("Productivity: %.1f\n", *p.ProductivityScore)
		}
		return nil
	},
}

var projectsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new project",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		title, _ := cmd.Flags().GetString("title")
		parent, _ := cmd.Flags().GetString("parent")
		color, _ := cmd.Flags().GetString("color")
		productivity, _ := cmd.Flags().GetFloat64("productivity")
		archived, _ := cmd.Flags().GetBool("archived")
		notes, _ := cmd.Flags().GetString("notes")
		billing, _ := cmd.Flags().GetString("billing-status")
		teamID, _ := cmd.Flags().GetString("team-id")

		payload := map[string]any{
			"title": title,
		}
		if parent != "" {
			payload["parent"] = parent
		}
		if color != "" {
			payload["color"] = color
		}
		if cmd.Flags().Changed("productivity") {
			payload["productivity_score"] = productivity
		}
		if cmd.Flags().Changed("archived") {
			payload["is_archived"] = archived
		}
		if notes != "" {
			payload["notes"] = notes
		}
		if billing != "" {
			payload["default_billing_status"] = billing
		}
		if teamID != "" {
			payload["team_id"] = teamID
		}

		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		body, err := c.Post("/projects", bytes.NewReader(data))
		if err != nil {
			return err
		}

		var resp struct {
			Data api.Project `json:"data"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if output.JSONOutput {
			output.PrintJSON(resp.Data)
			return nil
		}

		fmt.Printf("Created project: %s (%s)\n", resp.Data.Title, resp.Data.Self)
		return nil
	},
}

var projectsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()
		id := normalizeRef(args[0], "projects")

		payload := map[string]any{}
		if cmd.Flags().Changed("title") {
			v, _ := cmd.Flags().GetString("title")
			payload["title"] = v
		}
		if cmd.Flags().Changed("color") {
			v, _ := cmd.Flags().GetString("color")
			payload["color"] = v
		}
		if cmd.Flags().Changed("productivity") {
			v, _ := cmd.Flags().GetFloat64("productivity")
			payload["productivity_score"] = v
		}
		if cmd.Flags().Changed("archived") {
			v, _ := cmd.Flags().GetBool("archived")
			payload["is_archived"] = v
		}
		if cmd.Flags().Changed("notes") {
			v, _ := cmd.Flags().GetString("notes")
			payload["notes"] = v
		}
		if cmd.Flags().Changed("billing-status") {
			v, _ := cmd.Flags().GetString("billing-status")
			payload["default_billing_status"] = v
		}

		if len(payload) == 0 {
			return fmt.Errorf("no fields to update; use flags like --title, --color, --archived")
		}

		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		body, err := c.Put("/projects/"+id, bytes.NewReader(data))
		if err != nil {
			return err
		}

		var resp struct {
			Data api.Project `json:"data"`
		}
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}

		if output.JSONOutput {
			output.PrintJSON(resp.Data)
			return nil
		}

		fmt.Printf("Updated project: %s (%s)\n", resp.Data.Title, resp.Data.Self)
		return nil
	},
}

var projectsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()
		id := normalizeRef(args[0], "projects")

		if err := c.Delete("/projects/" + id); err != nil {
			return err
		}

		if output.JSONOutput {
			output.PrintJSON(map[string]string{"status": "deleted", "id": id})
		} else {
			fmt.Printf("Deleted project %s\n", id)
		}
		return nil
	},
}

// normalizeRef takes an input that may be a bare ID ("1") or a full reference
// ("/projects/1") and returns just the ID portion.
func normalizeRef(input, resource string) string {
	input = strings.TrimPrefix(input, "/"+resource+"/")
	input = strings.TrimPrefix(input, resource+"/")
	return input
}

func init() {
	projectsListCmd.Flags().Bool("tree", false, "show project hierarchy as a tree")
	projectsListCmd.Flags().String("title", "", "filter by title")
	projectsListCmd.Flags().Bool("hide-archived", false, "hide archived projects")
	projectsListCmd.Flags().String("team-id", "", "filter by team ID")

	projectsCreateCmd.Flags().String("title", "", "project title (required)")
	projectsCreateCmd.MarkFlagRequired("title")
	projectsCreateCmd.Flags().String("parent", "", "parent project reference (e.g. /projects/1)")
	projectsCreateCmd.Flags().String("color", "", "hex color (e.g. #FF0000)")
	projectsCreateCmd.Flags().Float64("productivity", 0, "productivity score (-1 to 1)")
	projectsCreateCmd.Flags().Bool("archived", false, "whether project is archived")
	projectsCreateCmd.Flags().String("notes", "", "project notes")
	projectsCreateCmd.Flags().String("billing-status", "", "default billing status")
	projectsCreateCmd.Flags().String("team-id", "", "team ID")

	projectsUpdateCmd.Flags().String("title", "", "project title")
	projectsUpdateCmd.Flags().String("color", "", "hex color")
	projectsUpdateCmd.Flags().Float64("productivity", 0, "productivity score (-1 to 1)")
	projectsUpdateCmd.Flags().Bool("archived", false, "whether project is archived")
	projectsUpdateCmd.Flags().String("notes", "", "project notes")
	projectsUpdateCmd.Flags().String("billing-status", "", "default billing status")

	projectsCmd.AddCommand(projectsListCmd)
	projectsCmd.AddCommand(projectsGetCmd)
	projectsCmd.AddCommand(projectsCreateCmd)
	projectsCmd.AddCommand(projectsUpdateCmd)
	projectsCmd.AddCommand(projectsDeleteCmd)
}
