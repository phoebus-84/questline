# Questline usage recipes

This page is a set of practical “recipes” for the Questline CLI/TUI.

## Quick flow

```bash
# Check status
ql status

# Add a quest
ql add "Buy groceries" --diff 2 --attr wis

# See your quest log
ql list

# Complete a quest
ql do 1
```

## Concepts (quick)

- **Task**: a quest you can complete for XP.
- **Project**: a container quest that groups subtasks (unlocks later).
- **Habit**: a recurring quest with cadence and diminishing returns (unlocks later).
- **Blueprint**: a gated template you can accept to spawn a predefined quest.

## Tasks

Create tasks with difficulty (1–5) and an attribute track:

```bash
ql add "Write one page" --diff 2 --attr int
ql add "Walk 20 minutes" --diff 1 --attr wis
```

## Subtasks

Subtasks unlock at a higher level. When unlocked:

```bash
# Create a parent
ql add "Read a Book" --diff 2 --attr art

# Add a child task under the parent
ql add "Chapter 1" --parent 12 --diff 2 --attr art
```

## Projects

Projects are containers (they aren’t completed directly; complete leaf tasks inside them):

```bash
ql add "Learn Go" --project --attr int
ql add "Read tour" --parent 12 --diff 2 --attr int
ql do 13
```

## Habits

Habits unlock at a higher level. When unlocked:

```bash
ql add "Push-ups" --habit --interval daily --diff 2 --attr str
```

## Blueprints

See blueprint availability in `ql status`, then accept one:

```bash
ql status
ql accept str_starter
```

If a blueprint is not available yet, Questline will tell you why.

## TUI dashboard

Open the dashboard:

```bash
ql board
```

Keymap:

- Move: `↑/↓` or `j/k`
- Expand/collapse: `enter`
- Complete: `c` or `space`
- Refresh: `r`
- Quit: `q`

## DB location

Default:

- `$HOME/.questline.db`

Override:

```bash
QL_DB_PATH=/tmp/questline.db ql status
```
