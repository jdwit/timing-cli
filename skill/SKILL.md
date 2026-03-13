---
name: timing-cli
description: "How to use the Timing CLI to manage time tracking: projects, time entries, timers, reports, and activity analysis. Use this skill whenever working with the timing CLI, the Timing app API, time tracking data, or when the user asks about their tracked time, projects, billing, productivity, or wants to log/query time entries. Also use when developing or modifying this CLI's codebase."
---

# Timing CLI

The `timing` CLI is a command-line interface for the Timing macOS time tracking app. Timing automatically records how users spend time across apps, websites, and documents in the background; the CLI lets you manage projects, create and query time entries, control timers, generate reports, and inspect the automatically-tracked activity hierarchy.

## Prerequisites

The CLI requires the `TIMING_API_KEY` environment variable to be set (a bearer token from https://web.timingapp.com/integrations/tokens). Optionally, `TIMING_TIMEZONE` overrides automatic timezone detection.

## Core concepts

Understanding these Timing concepts is essential for using the CLI effectively:

- **Projects** are hierarchical containers for time. They can be nested arbitrarily deep (e.g. "Work > Client A > Backend"). Each project has a color, productivity score (-1.0 to 1.0), and an optional default billing status. Projects can be archived to hide them without deleting data.
- **Time entries** are blocks of time assigned to a project with a title and optional notes. They have a billing status lifecycle: undetermined -> not_billable (terminal) OR billable -> billed -> paid. A running time entry is a timer.
- **Activities** are what Timing records automatically in the background: which apps, documents, and websites the user interacted with, and for how long. The activity hierarchy endpoint returns this data in a text format optimized for AI consumption, making it the most powerful tool for understanding how time was actually spent.
- **Reports** aggregate time entries with flexible grouping and filtering, useful for billing summaries, weekly reviews, and productivity analysis.

## Command reference

The CLI follows a `timing <resource> <action> [flags]` pattern. Append `--json` to any command for machine-readable output (essential for scripting and programmatic workflows).

Read `references/commands.md` for the complete flag-by-flag reference of every command.

### Quick reference

```
timing projects list [--tree] [--title X] [--hide-archived]
timing projects get <id>
timing projects create --title "Name" [--parent /projects/1] [--color #FF0000]
timing projects update <id> --title "New Name"
timing projects delete <id>

timing entries list [--start DATE] [--end DATE] [--project REF] [--search TEXT]
timing entries get <id>
timing entries create --start DATE --end DATE [--project REF] [--title TEXT]
timing entries update <id> [--title TEXT] [--project REF] [--notes TEXT]
timing entries delete <id>
timing entries batch-update --entries id1,id2 [--project REF] [--title TEXT]

timing timer start [--project REF] [--title TEXT]
timing timer stop
timing timer status

timing activities --start DATE --end DATE [--block-size total|day|hour|5min] [--max-lines N]

timing report [--start DATE] [--end DATE] [--project REF] [--columns project,title,timespan]
```

Command aliases: `projects`/`project`/`p`, `entries`/`entry`/`e`.

### ID handling

IDs are flexible: bare numbers (`1`), full references (`/projects/1`), or partial references (`projects/1`) all work. The CLI normalizes them automatically.

### Date formats

All dates use ISO 8601. For date-only values (e.g. `--start 2024-01-15`), the API interprets start as 00:00:00 and end as 23:59:59 in the configured timezone. For precise times, use the full format: `2024-01-15T09:30:00+01:00`.

## Leveraging the activity hierarchy

The `timing activities` command is the most powerful tool for AI agents. It returns Timing's automatically-tracked activity data as a tab-indented text hierarchy, designed specifically for LLM consumption. This data reveals exactly what apps, documents, and websites were used and for how long, even when the user hasn't created any manual time entries yet.

### Coarse-to-fine exploration

Start broad, then drill down. This prevents overwhelming context windows:

1. **Overview**: `timing activities --start 2024-01-15 --end 2024-01-15 --block-size total --max-lines 50`
   Shows top-level summary of where time went
2. **Hourly breakdown**: change `--block-size` to `hour` to see time distribution across the day
3. **Fine-grained**: use `5min` or `15min` to find exact boundaries for time entry creation

### Finding unassigned time

Use `--project-ids 0` to see only activities not yet assigned to any project. This is perfect for helping users categorize their uncategorized time.

### Practical workflows

**"What did I work on today?"**
```bash
timing activities --start 2024-01-15 --end 2024-01-15 --block-size hour --max-lines 200
```

**"Help me log my time for this week"**
1. Fetch the activity hierarchy for the date range
2. List existing projects with `timing projects list --tree`
3. Identify logical work blocks from the activity data
4. Create time entries for each block, assigning to the right project

**"How much billable time this month?"**
```bash
timing report --start 2024-01-01 --end 2024-01-31 \
  --billing-status billable,billed,paid \
  --columns project,title,timespan \
  --timespan-grouping week
```

## Ask, don't assume

When creating or updating time entries, ask the user for information they haven't provided rather than inventing defaults. If they say "log 2 hours on Backend" but don't mention a title or description, ask what to put there instead of making something up like "Work on Backend." The same applies to notes, billing status, and other optional fields: if the user cares enough to log time, they likely have an opinion about how it's labeled.

## Verify and check before creating entries

Timing has two layers of time data: automatically-tracked activities (matched to projects via rules) and manually created time entries. Before creating a manual entry, always verify against both layers to make sure the entry reflects reality.

### Step 1: Verify activities support the entry

The activity hierarchy is the ground truth of what actually happened. Before logging a block for a project, check that the tracked activities in that time range actually match:

```bash
timing activities --start DATE --end DATE --block-size hour --max-lines 200
```

Look at what's there:

- **Activities for the target project exist**: good, the manual block is justified. The activities are scattered fragments; the manual entry provides a clean accounting block over the same work.
- **Activities for other projects also exist in the range**: flag this to the user. For example: "There's 20 minutes of Zoncoalitie activity between 14:10-14:30 within your Gotit block. Do you want to log Zoncoalitie separately, adjust the Gotit block to exclude that period, or include it anyway?" This prevents accidentally burying work for one project inside a block claimed by another.
- **No activities for the target project exist**: question whether the block is correct. Maybe the user has the wrong time range or project.

This verification matters because Timing's Mac app uses rules to auto-assign activities to projects, but those are suggestions, not entries. The CLI can see these associations in the activity hierarchy (activities grouped under project names) but cannot see or interact with the Mac app's entry suggestions directly. The activity hierarchy is the best source of truth available.

### Step 2: Check for conflicting entries

List existing entries in the time range:

```bash
timing entries list --start START --end END --json
```

If entries overlap with the block you're about to create, inform the user. Explain what's there: which projects, titles, durations, and how they overlap. Then suggest options:

- **Replace**: create with `--replace-existing`, which trims partially overlapping entries and deletes fully contained ones. The user's manual block takes priority.
- **Skip**: leave existing data as-is.
- **Adjust**: modify the start/end times to avoid the conflict.
- **Delete then create**: remove specific conflicting entries first for full control.

### How `--replace-existing` handles partial overlaps

When a new entry partially overlaps an existing one, the existing entry gets trimmed (its start or end time is adjusted). When an existing entry falls entirely within the new block, it gets deleted. The user's manual blocks always win; anything outside them survives.

This fits a natural workflow: let Timing auto-track throughout the day, then create clean accounting blocks at review time. The scattered auto entries get trimmed around the clean blocks, and gaps keep their auto-assigned data.

## Important behaviors and gotchas

- **Starting a timer auto-stops any running timer.** There's no need to stop first.
- **Timer status returns "No timer running" gracefully** when nothing is active (the API returns 404, which the CLI handles).
- **The activity hierarchy is a beta endpoint** and returns plain text, not JSON. Even with `--json`, it wraps the text in a JSON object.
- **Date range limit**: the activity hierarchy endpoint has a maximum span of 32 days.
- **Default date range**: when no dates are specified on entries list, the API returns the last 30 days.
- **Rate limits**: 500 requests/hour, with throttling above 200 requests/minute. The CLI handles 429 responses with automatic retry (3 attempts, exponential backoff).
- **Project deletion is cascading**: deleting a project deletes all its children too.
- **PUT behaves like PATCH** for projects: omitted fields are not cleared.
- **Billing status inheritance**: entries inherit billing status from their project, which inherits from the team, which inherits from user settings; unless explicitly overridden.
- **`replace-existing` flag**: when creating entries or starting timers, this adjusts or deletes overlapping entries to prevent double-counting.
