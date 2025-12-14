# Contributing to Questline

Thanks for your interest in making Questline better.

Questline is a local-first RPG task manager written in Go (SQLite + Cobra CLI + Bubbletea TUI). Contributions should keep the project simple, fast, and pleasant.

## Ground rules

- Keep changes small and reviewable.
- Prefer clarity over cleverness.
- Don’t change gameplay/XP rules unless the issue explicitly calls for it.
- Run tests before you open a PR.

## Development setup

Requirements:

- Go 1.24+

Common commands:

```bash
# Run the CLI
go run ./cmd/ql --help

# Run tests
go test ./...

# Format
gofmt -w ./cmd ./internal
```

Database:

- Default DB path: `$HOME/.questline.db`
- Override via `QL_DB_PATH=/path/to/questline.db`

Example:

```bash
QL_DB_PATH=/tmp/questline-dev.db go run ./cmd/ql status
```

## Issue tracking (bd)

This repo uses **Beads** (`bd`) for all task tracking.

Workflow:

```bash
# Find unblocked work
bd ready --json

# Pick an issue and start it
bd update <id> --status in_progress --json

# Do the work, add tests/docs as needed

# Close when done
bd close <id> --reason "Done" --json

# Sync state to git (required)
bd sync
```

If you discover extra work while implementing something, create a linked issue:

```bash
bd create "Found: <short title>" -t bug -p 2 --deps discovered-from:<parent-id> --json
```

## Pull request checklist

- Tests pass: `go test ./...`
- Code is formatted: `gofmt -w ./cmd ./internal`
- You updated docs when behavior changed
- You updated/closed the relevant Beads issue and ran `bd sync`

## Style guide (quick)

- Keep CLI output stable and readable.
- Avoid new dependencies unless they clearly earn their keep.
- Keep functions small; prefer explicit names.

## Security

If you believe you’ve found a security issue, please open a Beads issue with minimal reproduction steps and avoid posting secrets.
