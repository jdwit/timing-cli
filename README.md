# timing-cli

CLI for the [Timing](https://timingapp.com) macOS time tracking app API.

## Installation

```bash
go install github.com/jdwit/timing-cli@latest
```

Or build from source:

```bash
git clone https://github.com/jdwit/timing-cli.git
cd timing-cli
go build -o timing .
```

## Configuration

Get your API key from the [Timing web app](https://web.timingapp.com) and set it as an
environment variable.

Add to your `~/.zshrc` (or `~/.bashrc`):

```bash
export TIMING_API_KEY="your-api-key"
export TIMING_TIMEZONE="Europe/Amsterdam"  # optional; defaults to system timezone
```

Then reload:

```bash
source ~/.zshrc
```

## Usage

All commands support `--json` for machine-readable output, making the CLI suitable for
scripting and AI agent integration.

### Projects

```bash
timing projects list                    # list all projects
timing projects list --tree             # hierarchical tree view
timing projects list --hide-archived    # exclude archived
timing projects get <id>                # show project details
timing projects create --title "Name" --color "#FF0000" --parent /projects/1
timing projects update <id> --title "New Name" --archived
timing projects delete <id>
```

### Time entries

```bash
timing entries list --start 2024-01-01 --end 2024-01-31
timing entries list --project /projects/1 --include-children
timing entries list --search "meeting" --billing-status billable
timing entries get <id>
timing entries create --start "2024-01-01T09:00:00+01:00" --end "2024-01-01T17:00:00+01:00" \
  --project /projects/1 --title "Work" --notes "Details"
timing entries update <id> --title "Updated" --notes "New notes"
timing entries delete <id>
timing entries batch-update --entries 1,2,3 --billing-status billable
```

### Timer

```bash
timing timer start --project /projects/1 --title "Working on feature"
timing timer status
timing timer stop
```

### Activities (for AI agents)

```bash
timing activities --start 2024-01-01 --end 2024-01-31
timing activities --start 2024-01-01 --end 2024-01-01 --block-size hour --max-lines 500
```

### Reports

```bash
timing report --start 2024-01-01 --end 2024-01-31
timing report --project /projects/1 --include-children --timespan-grouping day
timing report --billing-status billable --columns project,title,timespan
```

### Shell completion

```bash
# bash
source <(timing completion bash)

# zsh
timing completion zsh > "${fpath[1]}/_timing"
```

```bash
timing projects list --json
timing timer status --json
timing entries list --start 2024-01-01 --end 2024-01-31 --json
```

## AI agent skill

This repo includes an [agent skill](skill/) that teaches AI agents how to use the
Timing CLI effectively. It covers command usage, activity-based time verification,
conflict handling, and reporting workflows.
