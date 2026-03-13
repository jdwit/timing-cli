package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jdwit/timing-cli/internal/api"
	"github.com/jdwit/timing-cli/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func subcommandNames(cmd *cobra.Command) []string {
	var names []string
	for _, c := range cmd.Commands() {
		names = append(names, c.Name())
	}
	return names
}

func mockServer(t *testing.T, handler http.HandlerFunc) func() *api.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return func() *api.Client {
		c := api.NewClient("test-key", "UTC")
		c.BaseURL = srv.URL
		return c
	}
}

func resetFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		f.Changed = false
		if sv, ok := f.Value.(pflag.SliceValue); ok {
			sv.Replace(nil)
		} else {
			f.Value.Set(f.DefValue)
		}
	})
	for _, sub := range cmd.Commands() {
		resetFlags(sub)
	}
}

func executeCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()

	// Reset all flag state so tests don't bleed into each other
	jsonFlag = false
	output.JSONOutput = false
	resetFlags(rootCmd)

	// Capture stdout
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	rootCmd.SetArgs(args)
	execErr := rootCmd.Execute()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String(), execErr
}

// --- Root command ---

func TestRootCommand(t *testing.T) {
	assert.Equal(t, "timing", rootCmd.Use)
	assert.NotEmpty(t, rootCmd.Short)
}

func TestRootCommand_HasSubcommands(t *testing.T) {
	names := subcommandNames(rootCmd)
	for _, name := range []string{"projects", "entries", "timer", "activities", "report", "completion"} {
		assert.Contains(t, names, name, "root should have %q subcommand", name)
	}
}

func TestRootCommand_JSONFlag(t *testing.T) {
	f := rootCmd.PersistentFlags().Lookup("json")
	require.NotNil(t, f)
	assert.Equal(t, "false", f.DefValue)
}

// --- Projects ---

func TestProjectsCommand(t *testing.T) {
	assert.Equal(t, "projects", projectsCmd.Use)
	assert.Equal(t, []string{"project", "p"}, projectsCmd.Aliases)

	names := subcommandNames(projectsCmd)
	for _, name := range []string{"list", "get", "create", "update", "delete"} {
		assert.Contains(t, names, name)
	}
}

func TestProjectsListCmd_Flags(t *testing.T) {
	for _, name := range []string{"tree", "title", "hide-archived", "team-id"} {
		assert.NotNil(t, projectsListCmd.Flags().Lookup(name), "missing flag %q", name)
	}
}

func TestProjectsCreateCmd_Flags(t *testing.T) {
	for _, name := range []string{"title", "parent", "color", "productivity", "archived", "notes", "billing-status", "team-id"} {
		assert.NotNil(t, projectsCreateCmd.Flags().Lookup(name), "missing flag %q", name)
	}
}

func TestProjectsUpdateCmd_Flags(t *testing.T) {
	for _, name := range []string{"title", "color", "productivity", "archived", "notes", "billing-status"} {
		assert.NotNil(t, projectsUpdateCmd.Flags().Lookup(name), "missing flag %q", name)
	}
}

func TestProjectsGetCmd_RequiresArg(t *testing.T) {
	assert.NotNil(t, projectsGetCmd.Args)
}

func TestProjectsDeleteCmd_RequiresArg(t *testing.T) {
	assert.NotNil(t, projectsDeleteCmd.Args)
}

// --- Entries ---

func TestEntriesCommand(t *testing.T) {
	assert.Equal(t, "entries", entriesCmd.Use)
	assert.Equal(t, []string{"entry", "e"}, entriesCmd.Aliases)

	names := subcommandNames(entriesCmd)
	for _, name := range []string{"list", "get", "create", "update", "delete", "batch-update"} {
		assert.Contains(t, names, name)
	}
}

func TestEntriesListCmd_Flags(t *testing.T) {
	for _, name := range []string{"start", "end", "project", "search", "is-running", "include-children", "include-team-members", "billing-status", "page-limit"} {
		assert.NotNil(t, entriesListCmd.Flags().Lookup(name), "missing flag %q", name)
	}
}

func TestEntriesCreateCmd_Flags(t *testing.T) {
	for _, name := range []string{"start", "end", "project", "title", "notes", "replace-existing", "billing-status"} {
		assert.NotNil(t, entriesCreateCmd.Flags().Lookup(name), "missing flag %q", name)
	}
}

func TestEntriesUpdateCmd_Flags(t *testing.T) {
	for _, name := range []string{"start", "end", "project", "title", "notes", "replace-existing", "billing-status"} {
		assert.NotNil(t, entriesUpdateCmd.Flags().Lookup(name), "missing flag %q", name)
	}
}

func TestEntriesBatchUpdateCmd_Flags(t *testing.T) {
	for _, name := range []string{"entries", "project", "title", "notes", "billing-status", "allow-editing-other-users"} {
		assert.NotNil(t, entriesBatchUpdateCmd.Flags().Lookup(name), "missing flag %q", name)
	}
}

// --- Timer ---

func TestTimerCommand(t *testing.T) {
	assert.Equal(t, "timer", timerCmd.Use)

	names := subcommandNames(timerCmd)
	for _, name := range []string{"start", "stop", "status"} {
		assert.Contains(t, names, name)
	}
}

func TestTimerStartCmd_Flags(t *testing.T) {
	for _, name := range []string{"project", "title", "notes", "start", "replace-existing", "billing-status"} {
		assert.NotNil(t, timerStartCmd.Flags().Lookup(name), "missing flag %q", name)
	}
}

// --- Activities ---

func TestActivitiesCommand_Flags(t *testing.T) {
	assert.Equal(t, "activities", activitiesCmd.Use)
	for _, name := range []string{"start", "end", "block-size", "min-duration", "group-by-project", "max-depth", "max-lines", "include-mobile", "project-ids", "include-subprojects"} {
		assert.NotNil(t, activitiesCmd.Flags().Lookup(name), "missing flag %q", name)
	}
}

// --- Report ---

func TestReportCommand_Flags(t *testing.T) {
	assert.Equal(t, "report", reportCmd.Use)
	for _, name := range []string{"start", "end", "project", "include-children", "search", "columns", "project-grouping-level", "timespan-grouping", "billing-status", "sort", "include-app-usage", "include-team-members", "team-members"} {
		assert.NotNil(t, reportCmd.Flags().Lookup(name), "missing flag %q", name)
	}
}

// --- Completion ---

func TestCompletionCommand(t *testing.T) {
	assert.Equal(t, "completion [bash|zsh]", completionCmd.Use)
	assert.Equal(t, []string{"bash", "zsh"}, completionCmd.ValidArgs)
}

// --- Utility functions ---

func TestNormalizeRef(t *testing.T) {
	tests := []struct {
		input, resource, want string
	}{
		{"1", "projects", "1"},
		{"/projects/1", "projects", "1"},
		{"projects/1", "projects", "1"},
		{"/time-entries/42", "time-entries", "42"},
		{"99", "time-entries", "99"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, normalizeRef(tt.input, tt.resource), "normalizeRef(%q, %q)", tt.input, tt.resource)
	}
}

func TestFormatDate(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"2024-01-15T10:30:00+01:00", "2024-01-15T10:30"},
		{"2024-01-15T10:30", "2024-01-15T10:30"},
		{"short", "short"},
		{"", ""},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, formatDate(tt.input), "formatDate(%q)", tt.input)
	}
}

// --- Command execution tests ---

func TestProjectsList_JSON(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/projects", r.URL.Path)
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"self": "/projects/1", "title": "Test", "title_chain": []string{"Test"}, "color": "#FF0000"},
			},
		})
	})
	t.Cleanup(func() { newClient = orig })

	out, err := executeCommand(t, "projects", "list", "--json")
	require.NoError(t, err)

	var projects []api.Project
	require.NoError(t, json.Unmarshal([]byte(out), &projects))
	assert.Len(t, projects, 1)
	assert.Equal(t, "Test", projects[0].Title)
}

func TestProjectsList_Table(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"self": "/projects/1", "title": "Test", "title_chain": []string{"Test"}, "color": "#FF0000"},
			},
		})
	})
	t.Cleanup(func() { newClient = orig; output.JSONOutput = false })

	out, err := executeCommand(t, "projects", "list")
	require.NoError(t, err)
	assert.Contains(t, out, "Test")
}

func TestProjectsList_Tree(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/projects/hierarchy", r.URL.Path)
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"self": "/projects/1", "title": "Root", "children": []map[string]any{
					{"self": "/projects/2", "title": "Child"},
				}},
			},
		})
	})
	t.Cleanup(func() { newClient = orig; output.JSONOutput = false })

	out, err := executeCommand(t, "projects", "list", "--tree")
	require.NoError(t, err)
	assert.Contains(t, out, "Root")
	assert.Contains(t, out, "Child")
}

func TestProjectsGet_JSON(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/projects/1", r.URL.Path)
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"self": "/projects/1", "title": "Test", "title_chain": []string{"Test"}, "color": "#00F"},
		})
	})
	t.Cleanup(func() { newClient = orig })

	out, err := executeCommand(t, "projects", "get", "1", "--json")
	require.NoError(t, err)
	assert.Contains(t, out, "Test")
}

func TestProjectsGet_Table(t *testing.T) {
	orig := newClient
	score := 0.8
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"self": "/projects/1", "title": "Test", "title_chain": []string{"Test"},
				"color": "#00F", "is_archived": false, "default_billing_status": "billable",
				"parent": map[string]string{"self": "/projects/0"},
				"notes": "some notes", "productivity_score": score,
			},
		})
	})
	t.Cleanup(func() { newClient = orig; output.JSONOutput = false })

	out, err := executeCommand(t, "projects", "get", "1")
	require.NoError(t, err)
	assert.Contains(t, out, "Test")
	assert.Contains(t, out, "billable")
	assert.Contains(t, out, "some notes")
	assert.Contains(t, out, "Productivity")
}

func TestProjectsCreate_JSON(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		json.Unmarshal(body, &payload)
		assert.Equal(t, "New Project", payload["title"])

		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"self": "/projects/2", "title": "New Project"},
		})
	})
	t.Cleanup(func() { newClient = orig })

	out, err := executeCommand(t, "projects", "create", "--title", "New Project", "--json")
	require.NoError(t, err)
	assert.Contains(t, out, "New Project")
}

func TestProjectsCreate_Table(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"self": "/projects/2", "title": "New"},
		})
	})
	t.Cleanup(func() { newClient = orig; output.JSONOutput = false })

	out, err := executeCommand(t, "projects", "create", "--title", "New")
	require.NoError(t, err)
	assert.Contains(t, out, "Created project")
}

func TestProjectsUpdate_JSON(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"self": "/projects/1", "title": "Updated"},
		})
	})
	t.Cleanup(func() { newClient = orig })

	out, err := executeCommand(t, "projects", "update", "1", "--title", "Updated", "--json")
	require.NoError(t, err)
	assert.Contains(t, out, "Updated")
}

func TestProjectsUpdate_Table(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"self": "/projects/1", "title": "Updated"},
		})
	})
	t.Cleanup(func() { newClient = orig; output.JSONOutput = false })

	out, err := executeCommand(t, "projects", "update", "1", "--title", "Updated")
	require.NoError(t, err)
	assert.Contains(t, out, "Updated project")
}

func TestProjectsUpdate_NoFlags(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {})
	t.Cleanup(func() { newClient = orig; output.JSONOutput = false })

	_, err := executeCommand(t, "projects", "update", "1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no fields to update")
}

func TestProjectsDelete_JSON(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(204)
	})
	t.Cleanup(func() { newClient = orig })

	out, err := executeCommand(t, "projects", "delete", "1", "--json")
	require.NoError(t, err)
	assert.Contains(t, out, "deleted")
}

func TestProjectsDelete_Table(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	})
	t.Cleanup(func() { newClient = orig; output.JSONOutput = false })

	out, err := executeCommand(t, "projects", "delete", "1")
	require.NoError(t, err)
	assert.Contains(t, out, "Deleted project")
}

// --- Entries execution tests ---

func TestEntriesList_JSON(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"self": "/time-entries/1", "start_date": "2024-01-01T09:00:00", "end_date": "2024-01-01T10:00:00", "duration": 3600, "title": "Work"},
			},
			"meta":  map[string]int{"current_page": 1, "last_page": 1, "total": 1},
			"links": map[string]any{},
		})
	})
	t.Cleanup(func() { newClient = orig })

	out, err := executeCommand(t, "entries", "list", "--json")
	require.NoError(t, err)

	var entries []api.TimeEntry
	require.NoError(t, json.Unmarshal([]byte(out), &entries))
	assert.Len(t, entries, 1)
}

func TestEntriesList_Table(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{
					"self": "/time-entries/1", "start_date": "2024-01-01T09:00:00",
					"duration": 3600, "title": "Work", "is_running": false,
					"project": map[string]any{"self": "/projects/1", "title_chain": []string{"Dev"}},
				},
			},
			"meta":  map[string]int{"current_page": 1, "last_page": 1, "total": 1},
			"links": map[string]any{},
		})
	})
	t.Cleanup(func() { newClient = orig; output.JSONOutput = false })

	out, err := executeCommand(t, "entries", "list")
	require.NoError(t, err)
	assert.Contains(t, out, "Work")
}

func TestEntriesList_RunningEntry(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"self": "/time-entries/1", "start_date": "2024-01-01T09:00:00", "duration": 0, "is_running": true},
			},
			"meta":  map[string]int{"current_page": 1, "last_page": 1, "total": 1},
			"links": map[string]any{},
		})
	})
	t.Cleanup(func() { newClient = orig; output.JSONOutput = false })

	out, err := executeCommand(t, "entries", "list")
	require.NoError(t, err)
	assert.Contains(t, out, "running")
}

func TestEntriesGet_JSON(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"self": "/time-entries/1", "start_date": "2024-01-01T09:00:00",
				"end_date": "2024-01-01T10:00:00", "duration": 3600, "title": "Work",
			},
		})
	})
	t.Cleanup(func() { newClient = orig })

	out, err := executeCommand(t, "entries", "get", "1", "--json")
	require.NoError(t, err)
	assert.Contains(t, out, "Work")
}

func TestEntriesGet_Table(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"self": "/time-entries/1", "start_date": "2024-01-01T09:00:00",
				"end_date": "2024-01-01T10:00:00", "duration": 3600,
				"title": "Work", "notes": "some notes", "is_running": false,
				"billing_status": "billable",
				"project": map[string]any{"self": "/projects/1", "title_chain": []string{"Dev"}},
			},
		})
	})
	t.Cleanup(func() { newClient = orig; output.JSONOutput = false })

	out, err := executeCommand(t, "entries", "get", "1")
	require.NoError(t, err)
	assert.Contains(t, out, "Work")
	assert.Contains(t, out, "some notes")
	assert.Contains(t, out, "billable")
}

func TestEntriesCreate_JSON(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"self": "/time-entries/1", "duration": 3600},
		})
	})
	t.Cleanup(func() { newClient = orig })

	out, err := executeCommand(t, "entries", "create", "--start", "2024-01-01T09:00:00", "--end", "2024-01-01T10:00:00", "--json")
	require.NoError(t, err)
	assert.Contains(t, out, "/time-entries/1")
}

func TestEntriesCreate_Table(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"self": "/time-entries/1", "duration": 3600},
		})
	})
	t.Cleanup(func() { newClient = orig; output.JSONOutput = false })

	out, err := executeCommand(t, "entries", "create", "--start", "2024-01-01T09:00:00", "--end", "2024-01-01T10:00:00")
	require.NoError(t, err)
	assert.Contains(t, out, "Created time entry")
}

func TestEntriesUpdate_JSON(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"self": "/time-entries/1", "title": "Updated"},
		})
	})
	t.Cleanup(func() { newClient = orig })

	out, err := executeCommand(t, "entries", "update", "1", "--title", "Updated", "--json")
	require.NoError(t, err)
	assert.Contains(t, out, "Updated")
}

func TestEntriesUpdate_Table(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"self": "/time-entries/1"},
		})
	})
	t.Cleanup(func() { newClient = orig; output.JSONOutput = false })

	out, err := executeCommand(t, "entries", "update", "1", "--title", "Updated")
	require.NoError(t, err)
	assert.Contains(t, out, "Updated time entry")
}

func TestEntriesUpdate_NoFlags(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {})
	t.Cleanup(func() { newClient = orig; output.JSONOutput = false })

	_, err := executeCommand(t, "entries", "update", "1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no fields to update")
}

func TestEntriesDelete_JSON(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		w.WriteHeader(204)
	})
	t.Cleanup(func() { newClient = orig })

	out, err := executeCommand(t, "entries", "delete", "1", "--json")
	require.NoError(t, err)
	assert.Contains(t, out, "deleted")
}

func TestEntriesDelete_Table(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	})
	t.Cleanup(func() { newClient = orig; output.JSONOutput = false })

	out, err := executeCommand(t, "entries", "delete", "1")
	require.NoError(t, err)
	assert.Contains(t, out, "Deleted time entry")
}

func TestEntriesBatchUpdate_JSON(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"self": "/time-entries/1"},
				{"self": "/time-entries/2"},
			},
		})
	})
	t.Cleanup(func() { newClient = orig })

	out, err := executeCommand(t, "entries", "batch-update", "--entries", "1,2", "--title", "Batch", "--json")
	require.NoError(t, err)
	assert.Contains(t, out, "/time-entries/1")
}

func TestEntriesBatchUpdate_Table(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{{"self": "/time-entries/1"}, {"self": "/time-entries/2"}},
		})
	})
	t.Cleanup(func() { newClient = orig; output.JSONOutput = false })

	out, err := executeCommand(t, "entries", "batch-update", "--entries", "1,2", "--title", "Batch")
	require.NoError(t, err)
	assert.Contains(t, out, "Updated 2 time entries")
}

func TestEntriesBatchUpdate_NoEntries(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {})
	t.Cleanup(func() { newClient = orig; output.JSONOutput = false })

	_, err := executeCommand(t, "entries", "batch-update", "--title", "Batch")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--entries is required")
}

// --- Timer execution tests ---

func TestTimerStart_JSON(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"self": "/time-entries/1", "title": "Coding", "is_running": true},
		})
	})
	t.Cleanup(func() { newClient = orig })

	out, err := executeCommand(t, "timer", "start", "--title", "Coding", "--json")
	require.NoError(t, err)
	assert.Contains(t, out, "Coding")
}

func TestTimerStart_Table(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"self": "/time-entries/1", "title": "Coding"},
		})
	})
	t.Cleanup(func() { newClient = orig; output.JSONOutput = false })

	out, err := executeCommand(t, "timer", "start", "--title", "Coding")
	require.NoError(t, err)
	assert.Contains(t, out, "Timer started")
}

func TestTimerStart_Untitled(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"self": "/time-entries/1", "title": ""},
		})
	})
	t.Cleanup(func() { newClient = orig; output.JSONOutput = false })

	out, err := executeCommand(t, "timer", "start")
	require.NoError(t, err)
	assert.Contains(t, out, "(untitled)")
}

func TestTimerStart_WithMessage(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data":    map[string]any{"self": "/time-entries/1"},
			"message": "Timer already running",
		})
	})
	t.Cleanup(func() { newClient = orig; output.JSONOutput = false })

	out, err := executeCommand(t, "timer", "start")
	require.NoError(t, err)
	assert.Contains(t, out, "Timer already running")
}

func TestTimerStop_JSON(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"self": "/time-entries/1", "title": "Done", "duration": 3600},
		})
	})
	t.Cleanup(func() { newClient = orig })

	out, err := executeCommand(t, "timer", "stop", "--json")
	require.NoError(t, err)
	assert.Contains(t, out, "Done")
}

func TestTimerStop_Table(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{"self": "/time-entries/1", "title": "Done", "duration": 3600},
		})
	})
	t.Cleanup(func() { newClient = orig; output.JSONOutput = false })

	out, err := executeCommand(t, "timer", "stop")
	require.NoError(t, err)
	assert.Contains(t, out, "Timer stopped")
}

func TestTimerStatus_Running_JSON(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"self": "/time-entries/1", "title": "Working", "is_running": true,
				"start_date": "2024-01-01T09:00:00",
			},
		})
	})
	t.Cleanup(func() { newClient = orig })

	out, err := executeCommand(t, "timer", "status", "--json")
	require.NoError(t, err)
	assert.Contains(t, out, "Working")
}

func TestTimerStatus_Running_Table(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"self": "/time-entries/1", "title": "Working",
				"start_date": "2024-01-01T09:00:00",
				"project":    map[string]any{"self": "/projects/1", "title_chain": []string{"Dev"}},
			},
		})
	})
	t.Cleanup(func() { newClient = orig; output.JSONOutput = false })

	out, err := executeCommand(t, "timer", "status")
	require.NoError(t, err)
	assert.Contains(t, out, "Timer running")
	assert.Contains(t, out, "Working")
}

func TestTimerStatus_Untitled(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"self": "/time-entries/1", "title": "",
				"start_date": "2024-01-01T09:00:00",
			},
		})
	})
	t.Cleanup(func() { newClient = orig; output.JSONOutput = false })

	out, err := executeCommand(t, "timer", "status")
	require.NoError(t, err)
	assert.Contains(t, out, "(untitled)")
}

func TestTimerStatus_NotRunning_JSON(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})
	t.Cleanup(func() { newClient = orig })

	out, err := executeCommand(t, "timer", "status", "--json")
	require.NoError(t, err)
	assert.Contains(t, out, "false")
}

func TestTimerStatus_NotRunning_Table(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})
	t.Cleanup(func() { newClient = orig; output.JSONOutput = false })

	out, err := executeCommand(t, "timer", "status")
	require.NoError(t, err)
	assert.Contains(t, out, "No timer running")
}

// --- Activities ---

func TestActivities_JSON(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "text/plain", r.Header.Get("Accept"))
		w.Write([]byte("Activity data"))
	})
	t.Cleanup(func() { newClient = orig })

	out, err := executeCommand(t, "activities", "--json")
	require.NoError(t, err)
	assert.Contains(t, out, "Activity data")
}

func TestActivities_Table(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Activity data\n"))
	})
	t.Cleanup(func() { newClient = orig; output.JSONOutput = false })

	out, err := executeCommand(t, "activities")
	require.NoError(t, err)
	assert.Contains(t, out, "Activity data")
}

// --- Teams ---

// --- Report ---

func TestReport_JSON(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"duration": 3600, "title": "Coding", "project": map[string]any{"self": "/projects/1", "title_chain": []string{"Dev"}}},
			},
		})
	})
	t.Cleanup(func() { newClient = orig })

	out, err := executeCommand(t, "report", "--json")
	require.NoError(t, err)
	assert.Contains(t, out, "Coding")
}

func TestReport_Table(t *testing.T) {
	orig := newClient
	newClient = mockServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{
				{"duration": 3600, "title": "Coding", "project": map[string]any{"self": "/projects/1", "title_chain": []string{"Dev"}}},
			},
		})
	})
	t.Cleanup(func() { newClient = orig; output.JSONOutput = false })

	out, err := executeCommand(t, "report")
	require.NoError(t, err)
	assert.Contains(t, out, "Coding")
	assert.Contains(t, out, "Total")
}

// --- Completion ---

func TestCompletion_Bash(t *testing.T) {
	out, err := executeCommand(t, "completion", "bash")
	require.NoError(t, err)
	assert.Contains(t, out, "bash")
}

func TestCompletion_Zsh(t *testing.T) {
	out, err := executeCommand(t, "completion", "zsh")
	require.NoError(t, err)
	assert.NotEmpty(t, out)
}

func TestCompletion_InvalidShell(t *testing.T) {
	_, err := executeCommand(t, "completion", "invalid")
	require.Error(t, err)
}
