---
name: git-context
description: Store and manage coding context in git. Use when saving decisions, tracking tasks, planning complex work, or coordinating with other agents. Activates on mentions of "save context", "track task", "planning", or multi-agent coordination.
---

# git-context Skill

Use `git ctx` to persist context in git. Entries live in `.git/context` and are shared by every worktree of the clone. Shared storage that syncs with the team (`--shared`, `push`, `pull`) is not available yet and exits with an error.

## Quick Reference

```bash
# Memory (context entries)
git ctx add --title "Title" -m "Content"   # Add context
git ctx add --title "Title"                 # Opens editor
git ctx list                                # List entries, newest first
git ctx show <id>                           # View entry
git ctx edit <id>                           # Edit entry
git ctx rm <id>                             # Remove
git ctx search "query"                      # Search

# Tasks
git ctx task add "Title"                    # Create task
git ctx task add "Title" -d "Description"   # With description
git ctx task list                           # List tasks
git ctx task show <id>                      # View task
git ctx task claim <id>                     # Claim (take ownership; safe to retry)
git ctx task drop <id>                      # Release ownership
git ctx task done <id>                      # Complete
git ctx task comment <id> "message"         # Add comment

# Locks (atomic, expire after 4 hours)
git ctx lock <path-or-task>                 # Lock; fails if someone else holds it
git ctx lock list                           # Active locks
git ctx unlock <path-or-task>               # Release one lock
git ctx unlock                              # Release all your locks, clear expired ones

# Flags
--json                                      # JSON output (list, show, search, task list, task show, lock list)
```

Non-interactive sessions must pass `-m` or pipe content on stdin; `git ctx add` without either opens an editor and fails when there is no terminal.

## When to Use

### Save Context When:
- Making **architecture decisions** → `git ctx add --title "Why PostgreSQL"`
- Discovering **important information** → `git ctx add --title "API rate limits"`
- **Ending a session** → `git ctx add --title "Session handoff"`
- Finding **gotchas or bugs** → `git ctx add --title "Bug: Auth edge case"`

### Use Tasks When:
- **Breaking down** complex work
- **Coordinating** with other agents
- Work needs **tracking** across sessions

### Use Locks When:
- Several agents may edit the **same path** at once
- Several agents may race to **claim the same task** — lock the task id first

## Planning Pattern (Manus-style)

For complex tasks, create a plan entry:

```bash
git ctx add --title "Plan: [Feature Name]" << 'EOF'
## Goal
[One sentence describing success]

## Phases
- [ ] Phase 1: Setup and planning
- [ ] Phase 2: Core implementation
- [ ] Phase 3: Testing
- [ ] Phase 4: Documentation

## Key Questions
1. [Question to answer]
2. [Question to answer]

## Decisions Made
- (none yet)

## Errors Encountered
- (none yet)

## Status
**Currently in Phase 1** - Planning
EOF
```

### The Loop

1. **Before each major decision** → Re-read the plan
   ```bash
   git ctx show <plan-id>
   ```

2. **After each phase** → Update the plan
   ```bash
   git ctx edit <plan-id>
   # Mark [x] completed, update Status section
   ```

3. **When you learn something** → Save to separate entry
   ```bash
   git ctx add --title "Notes: [Topic]"
   ```

This keeps goals in your attention window and builds knowledge.

## Multi-Agent Workflow

Agents in worktrees of the same clone see the same tasks and locks.

### As Team Lead
```bash
git ctx task add "Implement auth" -d "JWT with refresh tokens"
git ctx task add "Setup database" -d "PostgreSQL schema"
git ctx task add "Write tests"
```

### As Worker Agent
```bash
# 1. Check available tasks
git ctx task list
# task-001  Implement auth   [open]
# task-002  Setup database   [open]

# 2. Claim one (lock it first if other agents may race for it)
git ctx lock task-001
git ctx task claim task-001

# 3. Lock the paths you will edit
git ctx lock src/auth/

# 4. Work on it, add comments
git ctx task comment task-001 "Using bcrypt for passwords"

# 5. Complete and release
git ctx task done task-001
git ctx unlock
```

### Avoiding Conflicts
- **Lock before editing** a path other agents may touch; `lock` tells you who holds it and until when
- **Claim before working**; a claim by someone else is refused
- **Release with `git ctx unlock`** when done; locks expire after 4 hours anyway

## Session Handoff

At end of session:

```bash
git ctx add --title "Handoff: [Date]" << 'EOF'
## Completed
- [What was done]

## In Progress
- [Current state]

## Next Steps
- [What to do next]

## Blockers
- [Any issues]
EOF
```

Next session:
```bash
git ctx list
git ctx show <handoff-id>
```

## Anti-Patterns

| Don't | Do Instead |
|-------|------------|
| Forget context between sessions | Save with `git ctx add` |
| Start complex work immediately | Create plan first |
| Edit a shared path without locking | `git ctx lock <path>` first |
| Keep findings in head | Save to entries |
| Retry errors silently | Log errors in plan |

## Storage

```
.git/
└── context/              ← LOCAL (private to the clone, shared by its worktrees)
    ├── memory/
    ├── tasks/
    └── locks/
```

Shared storage under `refs/context/` that syncs with push/pull is planned.

## Tips

1. **IDs are short** - Use the full 8 chars: `git ctx show abc12345`
2. **Pipe content** - `echo "text" | git ctx add --title "Note"`
3. **Search is fast** - `git ctx search "auth"` finds all related
4. **JSON for scripts** - `git ctx list --json | jq ...`
