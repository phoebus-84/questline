# Questline

Questline is a local-first RPG task manager built in Go.

You add quests, complete them for XP, and unlock mechanics as your character grows — all stored in a tiny local SQLite file.

**Vibe:** sleek and playful, still serious about ergonomics.

## Features

- Local-first SQLite DB (single file)
- RPG progression: XP, levels, gates/unlocks
- Tasks, projects, subtasks, and recurring habits
- Blueprints (unlockable templates)
- CLI + Bubbletea TUI dashboard

## Requirements

- Go 1.24+

## Build & Run

```bash
# Run from source
go run ./cmd/ql --help

# Build a binary
go build ./cmd/ql

# Or install to your GOPATH/bin
go install ./cmd/ql
```

## Docs

- Usage recipes: [docs/USAGE.md](docs/USAGE.md)
- Contributing: [CONTRIBUTING.md](CONTRIBUTING.md)
- License: [LICENSE](LICENSE)

## Quickstart (5 steps)

```bash
# 1) Build a local binary
go build -o ql ./cmd/ql

# 2) Check your status (creates the DB on first run)
./ql status

# 3) Add your first quest
./ql add "Buy groceries" --diff 2 --attr wis

# 4) View your quest log
./ql list

# 5) Complete a quest for XP
./ql do 1
```

Optional:

```bash
# Open the TUI dashboard
./ql board
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

# Add a habit with duration and goal
ql add "Learn Spanish" --habit --interval daily --duration 30d --goal 30 --attr int

# Complete a task/habit by ID
ql do 42

# Restore (undo) a completion
ql restore 42

# Print a tree view
ql list

# Accept a blueprint (once available)
ql accept str_starter

# Open the TUI dashboard
ql board
```

Notes:

- `--diff` is 1–5 (trivial → epic).
- Attributes: `str|int|wis|art|home|out|read|cinema|career`.

## Sample outputs

These examples show the *shape* of the output (colors may vary by terminal).

### `ql status`

```text
✨ Player Status
Level: 0
Total XP: 0 (next at 500, 500 to go)

📊 Attributes
- 💪 STR: lvl 0 (xp 0)
- 🧠 INT: lvl 0 (xp 0)
- 🧘 WIS: lvl 0 (xp 0)
- 🎨 ART: lvl 0 (xp 0)
- 🏠 HOME: lvl 0 (xp 0)
- 🌳 OUT: lvl 0 (xp 0)
- 📚 READ: lvl 0 (xp 0)
- 🎬 CINEMA: lvl 0 (xp 0)
- 💼 CAREER: lvl 0 (xp 0)

🔓 Gates
- Max active tasks: 3 (currently 0)
- Subtasks: locked
- Habits: locked
- Projects: locked

🔒 Blueprints (locked):
- art_critic
- art_reader
- str_starter
```

### `ql add "..."`

```text
➕ Created task 🗺️ #1 Buy groceries (+100 XP)
```

### `ql list`

```text
🗺️ Quest Log
🗺️ #1 Buy groceries (pending)
```

### `ql do <id>`

```text
✅ Completed 🗺️ #1 Buy groceries (+100 XP)
Level: 0 → 0
```

### `ql accept <blueprint_id>`

If the blueprint is still locked:

```text
🧨 blueprint str_starter is not available (status=locked)
```

Once you reach the required gate (e.g. habits unlock at level 5):

```text
📜 Accepted str_starter → created #8
```

### `ql board`

`ql board` opens an interactive dashboard (TUI). It doesn’t print a static report; use the keymap below.

## TUI (`ql board`)

The TUI is a retro gaming-style dashboard with phosphor green/amber aesthetic:

- **Header**: Level, total XP, progress to next level
- **Sidebar**: Attribute levels, weekly XP spark graphs, achievements/badges, key help
- **Main**: Focus list (suggested tasks) + collapsible quest log tree
- **Footer**: Status messages and tips

### Keybindings

| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate |
| `Enter` | Expand/collapse |
| `c` or `Space` | Complete task |
| `d` or `Backspace` | Delete task (with confirmation) |
| `r` | Refresh |
| `?` | Toggle help |
| `q` | Quit |

## Issue Tracking (Beads)

This repo uses Beads (`bd`) for all issue tracking.

```bash
bd ready --json
bd update <id> --status in_progress --json
bd close <id> --reason "Done" --json
bd sync
```
