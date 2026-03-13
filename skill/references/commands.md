# Timing CLI: Complete Command Reference

## Global flags

| Flag | Description |
|------|-------------|
| `--json` | Output in JSON format (works on every command) |

## projects (aliases: project, p)

### projects list

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--tree` | bool | false | Show hierarchy as indented tree (uses `/projects/hierarchy` endpoint) |
| `--title` | string | | Filter by title (word search) |
| `--hide-archived` | bool | false | Hide archived projects and their children |
| `--team-id` | string | | Filter by team ID |

### projects get \<id\>

No additional flags. Accepts bare ID, `/projects/1`, or `projects/1`.

### projects create

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--title` | string | yes | Project title |
| `--parent` | string | | Parent project reference (e.g. `/projects/1`) |
| `--color` | string | | Hex color (e.g. `#FF0000`) |
| `--productivity` | float64 | | Productivity score (-1.0 to 1.0) |
| `--archived` | bool | | Whether project is archived |
| `--notes` | string | | Project notes |
| `--billing-status` | string | | Default billing status |
| `--team-id` | string | | Team ID |

### projects update \<id\>

Same flags as create (except `--title` and `--team-id` are optional). At least one flag must be provided.

### projects delete \<id\>

No additional flags. Cascading: deletes all child projects too.

## entries (aliases: entry, e)

### entries list

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--start` | string | | Start date filter (ISO 8601) |
| `--end` | string | | End date filter (ISO 8601) |
| `--project` | string | | Filter by project reference |
| `--search` | string | | Search titles and notes |
| `--is-running` | string | | Filter by running state (`true`/`false`) |
| `--include-children` | bool | false | Include child project entries |
| `--include-team-members` | bool | false | Include team members' entries |
| `--billing-status` | []string | | Filter by billing status (repeatable) |
| `--page-limit` | int | 0 | Max pages to fetch (0 = all) |

### entries get \<id\>

No additional flags.

### entries create

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--start` | string | yes | Start date/time (ISO 8601) |
| `--end` | string | yes | End date/time (ISO 8601) |
| `--project` | string | | Project reference |
| `--title` | string | | Entry title |
| `--notes` | string | | Entry notes |
| `--replace-existing` | bool | | Replace overlapping entries |
| `--billing-status` | string | | Billing status |

### entries update \<id\>

Same flags as create (all optional). At least one flag must be provided.

### entries delete \<id\>

No additional flags.

### entries batch-update

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--entries` | []string | yes | Comma-separated entry IDs |
| `--project` | string | | Project reference |
| `--title` | string | | Entry title |
| `--notes` | string | | Entry notes |
| `--billing-status` | string | | Billing status |
| `--allow-editing-other-users` | bool | | Allow editing other users' entries |

## timer

### timer start

| Flag | Type | Description |
|------|------|-------------|
| `--project` | string | Project reference |
| `--title` | string | Timer title |
| `--notes` | string | Timer notes |
| `--start` | string | Custom start time (ISO 8601) |
| `--replace-existing` | bool | Replace overlapping entries |
| `--billing-status` | string | Billing status |

Starting a timer automatically stops any currently running timer.

### timer stop

No flags. Stops the running timer and prints its duration.

### timer status

No flags. Shows the running timer or "No timer running."

## activities

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--start` | string | | Start date (required, date-only e.g. `2024-01-01`) |
| `--end` | string | | End date (required, date-only e.g. `2024-01-31`) |
| `--block-size` | string | `total` | Granularity: `total`, `month`, `week`, `day`, `hour`, `15min`, `5min` |
| `--min-duration` | int | 60 | Minimum activity duration in seconds |
| `--group-by-project` | bool | true | Group by project |
| `--max-depth` | int | 0 | Maximum nesting depth (0 = unlimited) |
| `--max-lines` | int | 100 | Maximum leaf activities (1-1000) |
| `--include-mobile` | bool | false | Include iPhone/iPad data |
| `--project-ids` | []string | | Filter to project IDs (use `0` for unassigned) |
| `--include-subprojects` | bool | true | Include child projects |

Returns plain text (tab-indented hierarchy), not JSON. Duration format: `H:MM:SS`. Max date range: 32 days.

## report

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--start` | string | | Start date (ISO 8601) |
| `--end` | string | | End date (ISO 8601) |
| `--project` | []string | | Filter by project references |
| `--include-children` | bool | false | Include child project entries |
| `--search` | string | | Search entry titles/notes |
| `--columns` | []string | | Columns: `project`, `title`, `notes`, `timespan`, `user`, `billing_status` |
| `--project-grouping-level` | int | -1 | Aggregate at hierarchy level (-1 = no grouping, 0 = top-level) |
| `--timespan-grouping` | string | | Grouping: `exact`, `day`, `week`, `month`, `year` |
| `--billing-status` | []string | | Filter by billing status |
| `--sort` | []string | | Sort columns (prefix `-` for descending, default: `-duration`) |
| `--include-app-usage` | bool | false | Include app usage (server-intensive) |
| `--include-team-members` | bool | false | Include team members |
| `--team-members` | []string | | Restrict to specific team members |

## completion

Generate shell completions:
```
timing completion bash
timing completion zsh
timing completion fish
timing completion powershell
```
