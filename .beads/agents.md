# Beads Agent Examples

This guide shows how AI agents can use Beads for issue tracking.

## Basic Agent Workflows

### 1. Creating Issues from Code Review

When an agent identifies problems during code review:

```bash
# Create issue for bug found
bd create "Fix null pointer in user.login()" --priority high --labels bug

# Create issue for improvement
bd create "Refactor authentication module" --labels enhancement

# Create issue with description
bd create "Add error handling" --description "Need to handle network timeouts in API calls"
```

### 2. Working on Issues

When an agent picks up work:

```bash
# List available issues
bd list --status open

# Start working on an issue
bd update issue-1 --status in_progress --assign @agent

# Add progress notes
bd comment issue-1 "Implemented basic error handling, testing edge cases"

# Mark complete
bd update issue-1 --status done
```

### 3. Organizing Work

Agents can organize issues by priority and labels:

```bash
# View high priority items
bd list --priority high

# View bugs only
bd list --labels bug

# View in-progress work
bd list --status in_progress

# Filter by multiple criteria
bd list --status open --priority high --labels bug
```

### 4. Syncing with Team

After making changes:

```bash
# Sync issues with remote
bd sync

# View recent changes
bd log --limit 10
```

## Agent Integration Patterns

### Pattern 1: Issue-Driven Development

1. Agent reviews codebase and creates issues
2. Agent prioritizes issues
3. Agent works through issues one by one
4. Agent updates status and adds comments
5. Agent syncs when done

### Pattern 2: Collaborative Workflow

1. Multiple agents share the same issue database
2. Each agent claims issues by assigning to themselves
3. Agents comment on issues to coordinate
4. Regular syncing keeps everyone updated

### Pattern 3: Automated Triage

1. Agent scans for patterns (TODOs, FIXMEs, etc.)
2. Agent creates issues automatically
3. Agent labels and prioritizes based on context
4. Agent notifies team via sync

## Common Commands Quick Reference

```bash
# CRUD Operations
bd create "<title>"                    # Create issue
bd show <issue-id>                     # View details
bd update <issue-id> --status <status> # Update issue
bd delete <issue-id>                   # Delete issue

# Status Management
--status open | in_progress | done | blocked

# Priority Levels  
--priority low | medium | high | critical

# Common Labels
--labels bug,enhancement,documentation,testing

# Queries
bd list                                # All issues
bd list --status open                  # Open issues only
bd list --priority high                # High priority
bd list --labels bug                   # Bugs only
bd list --assign @agent                # Assigned to agent

# Collaboration
bd comment <issue-id> "<message>"      # Add comment
bd sync                                # Sync with remote
bd log                                 # View change history
```

## Example: Complete Agent Session

```bash
# 1. Start work session - sync latest
bd sync

# 2. Review what needs to be done
bd list --status open --priority high

# 3. Pick an issue and start working
bd update issue-1 --status in_progress --assign @copilot

# 4. Make code changes...
# (agent implements the feature)

# 5. Document progress
bd comment issue-1 "Implemented user authentication with OAuth2. Tests passing."

# 6. Mark complete
bd update issue-1 --status done

# 7. Sync changes
bd sync
```

## Tips for AI Agents

1. **Always sync first**: Start each session with `bd sync` to get latest issues
2. **Use clear titles**: Make issue titles specific and actionable
3. **Add context**: Use descriptions and comments to explain decisions
4. **Tag appropriately**: Use labels and priority to organize work
5. **Update status**: Keep status current so others know what's being worked on
6. **Comment progress**: Regular comments help track what was tried and why
7. **Sync regularly**: Push your issue updates so the team stays informed

## Advanced Features

### Linking Issues to Commits

When committing code that addresses an issue:

```bash
git commit -m "Fix login bug

Addresses: issue-1
- Added null check before user.login()
- Added unit tests for edge cases"
```

### Bulk Operations

```bash
# Close multiple completed issues
bd list --status done | xargs -n1 bd update --status closed

# Add label to multiple issues
bd list --status open | xargs -n1 bd update --labels needs-review
```

### Issue Dependencies

Use comments to track dependencies:

```bash
bd comment issue-2 "Blocked by: issue-1 (need auth before implementing this)"
bd update issue-2 --status blocked
```

---

**Note**: This repository is using Beads v1.x. For full documentation, see: https://github.com/steveyegge/beads
