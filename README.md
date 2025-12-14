# Organizer

A task management application with AI-native issue tracking using Beads.

## Getting Started

### Prerequisites

- Python 3.8+
- Beads CLI installed (`go install github.com/steveyegge/beads/cmd/bd@latest`)

### Installation

```bash
# Install dependencies
pip install -r requirements.txt

# Run the application
python src/main.py
```

## Issue Tracking

This project uses [Beads](https://github.com/steveyegge/beads) for issue tracking. Issues are stored in `.beads/issues.jsonl` and synced via git.

### Quick Commands

```bash
# Create a new issue
bd create "Your issue title"

# List all issues
bd list

# View issue details
bd show issue-1

# Update issue status
bd update issue-1 --status in_progress

# Sync with remote
bd sync
```

For more details on using Beads with AI agents, see [`.beads/agents.md`](.beads/agents.md).

## Development

This project is designed to work seamlessly with AI coding assistants. Issues can be created, updated, and tracked directly from the command line.

## License

MIT
