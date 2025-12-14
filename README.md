# Questline (MVP)

Questline is a local-first RPG task manager built in Go. It stores state in a local SQLite DB and provides a CLI (and a minimal TUI) for creating tasks, completing them for XP, and unlocking features as you level up.

## Requirements

- Go 1.24+

## Build & Run

```bash
# Run from source
go run ./cmd/ql -- --help

# Build a binary
go build ./cmd/ql

# Or install to your GOPATH/bin
go install ./cmd/ql@latest
```

## Database Location

By default Questline uses:

- `$HOME/.questline.db`

Override with:

- `QL_DB_PATH=/path/to/questline.db`

Example:

```bash
QL_DB_PATH=/tmp/questline.db go run ./cmd/ql -- status
```

## CLI Cheatsheet

```bash
# Show stats, gates, blueprint availability
ql status

# Add a task
ql add "Buy groceries" --diff 2 --attr wis

# Add a project container (requires project unlock)
ql add "Read a Book" --project --attr art

# Add a subtask under a parent
ql add "Chapter 1" --parent 12 --diff 2 --attr art

# Add a habit (requires habit unlock)
ql add "Push-ups" --habit --interval daily --diff 2 --attr str

# Complete a task/habit by ID
ql do 42

# Print a tree view
ql list

# Accept a blueprint (once available)
ql accept str_starter

# Open the TUI dashboard
ql board
```

Notes:

- `--diff` is 1–5 (trivial → epic).
- Attributes: `str|int|wis|art`.

## TUI (`ql board`)

The TUI is a minimal Bubbletea dashboard:

- Sidebar: attribute progress + keys
- Main: focus list (1–3 suggested leaf tasks) + collapsible quest log tree
- Actions: `↑/↓` (or `j/k`), `enter` (expand/collapse), `c`/space (complete), `r` (refresh), `q` (quit)

## Issue Tracking (Beads)

This repo uses Beads (`bd`) for all issue tracking.

```bash
bd ready --json
bd update <id> --status in_progress --json
bd close <id> --reason "Done" --json
bd sync
```

## Legacy Python Placeholder

There is an older Python placeholder app in `src/` which is not used by the Questline Go implementation.
