# git-context

**Offline-first context storage embedded in git, with multi-agent support.**

Store coding context, decisions, tasks and locks directly in your git repository. Everything lives in `.git/`, works offline, and is shared by every worktree of a clone. Designed for AI-assisted development and multi-agent collaboration.

## Features

- **Embedded in git**: Context lives in `.git/context/` — no external database, no server
- **Offline-first**: Works without network
- **Multi-agent ready**: Tasks, atomic locks, and coordination primitives for agents working in the same clone
- **Markdown-first**: Human-readable context entries
- **Planned**: Shared storage in `refs/context/` that syncs with the team via `git ctx push/pull`

## Installation

### Homebrew (macOS)

```bash
brew tap jxucoder/tap
brew install git-ctx
git config --global alias.ctx '!git-ctx'
```

### Download Binary

Download the appropriate binary from [Releases](https://github.com/jxucoder/git-context/releases):

| Platform | Binary |
|----------|--------|
| macOS (Apple Silicon) | `git-ctx-darwin-arm64` |
| macOS (Intel) | `git-ctx-darwin-amd64` |
| Linux (x64) | `git-ctx-linux-amd64` |
| Linux (ARM64) | `git-ctx-linux-arm64` |
| Windows (x64) | `git-ctx-windows-amd64.exe` |

```bash
# Example: macOS Apple Silicon
curl -L -o git-ctx https://github.com/jxucoder/git-context/releases/download/v0.3.1/git-ctx-darwin-arm64
chmod +x git-ctx
mv git-ctx ~/.local/bin/
git config --global alias.ctx '!git-ctx'
```

### Go

```bash
go install github.com/jxucoder/git-context/cmd/git-ctx@latest
git config --global alias.ctx '!git-ctx'
```

### Build from Source

```bash
make install
```

## Quick Start

```bash
# Add context
git ctx add --title "Why JWT for auth"
# Opens your editor, or use: git ctx add -t "Title" -m "Content"

# List entries (newest first)
git ctx list

# Show an entry
git ctx show abc123

# Search
git ctx search "auth"
```

## Tasks (Multi-Agent)

```bash
# Create tasks
git ctx task add "Implement user auth"

# List and claim
git ctx task list
git ctx task claim task-abc123     # Claiming a task you already own is a no-op

# Release or complete
git ctx task drop task-abc123
git ctx task done task-abc123

# Add comments
git ctx task comment task-abc123 "Using bcrypt for passwords"
```

## Locks

Locks tell other agents to keep off a path or task. They are acquired atomically, expire after 4 hours, and are visible from every worktree of the clone.

```bash
git ctx lock src/auth/       # src/auth, src/auth/ and ./src/auth are the same lock
git ctx lock task-abc123     # Lock a task
git ctx lock list            # Active locks, expiry shown in local time
git ctx unlock src/auth/     # Release one lock (expired locks can be released by anyone)
git ctx unlock               # Release all your locks and clear expired ones
```

## Storage Model

Context is stored inside `.git/`, keeping your working directory clean.

| Mode | Location | Status |
|------|----------|--------|
| Local (default) | `.git/context/` | Available. Private to the clone, shared by all of its worktrees |
| Shared | `.git/refs/context/` | Planned. `--shared`, `push` and `pull` currently exit with "not implemented" |

## Commands

### Memory (Context Entries)

| Command | Description |
|---------|-------------|
| `git ctx add [--title "T"] [-m "content"]` | Add entry (opens your editor when no content is given) |
| `git ctx list` | List entries, newest first |
| `git ctx show <id>` | View entry |
| `git ctx edit <id>` | Edit entry |
| `git ctx rm <id>` | Remove entry |
| `git ctx search "query"` | Search entries |

### Tasks

| Command | Description |
|---------|-------------|
| `git ctx task add "title" [-d "description"]` | Create task |
| `git ctx task list` | List tasks, newest first |
| `git ctx task show <id>` | View task details |
| `git ctx task claim <id>` | Take ownership |
| `git ctx task drop <id>` | Release ownership |
| `git ctx task done <id>` | Mark complete |
| `git ctx task comment <id> "msg"` | Add comment |

### Locks

| Command | Description |
|---------|-------------|
| `git ctx lock <target>` | Lock a path or task |
| `git ctx lock list` | List active locks |
| `git ctx unlock [target]` | Release a lock, or all your locks |

### Sync (planned)

| Command | Description |
|---------|-------------|
| `git ctx push` | Push shared entries to remote — not available yet |
| `git ctx pull` | Pull shared entries from remote — not available yet |

### Flags

| Flag | Description |
|------|-------------|
| `--json` | Output as JSON (`list`, `show`, `search`, `task list`, `task show`, `lock list`) |
| `--all`, `-a` | Show both local and shared (shared is empty until implemented) |
| `--shared`, `-s` | Use shared storage — not available yet, exits with an error |

Editors are resolved like git: `GIT_EDITOR`, then `core.editor`, `VISUAL`, `EDITOR`, falling back to `vim`. Values with arguments such as `code --wait` work.

## Multi-Agent Workflow

Agents that work in worktrees of the same clone share tasks and locks.

1. **Lead creates tasks:**
   ```bash
   git ctx task add "Implement auth"
   git ctx task add "Setup database"
   ```

2. **Agents claim work and lock what they touch:**
   ```bash
   git ctx task list
   git ctx task claim task-abc123
   git ctx lock src/auth/
   ```

3. **Agents complete and release:**
   ```bash
   git ctx task done task-abc123
   git ctx unlock
   ```

Locks are acquired atomically: when two agents race for the same path, exactly one wins and the other is told who holds it and until when. Lock a task before claiming it when several agents may race for the same task.

## Use with Claude (AI Skill)

Copy the skill to your Claude skills directory:

```bash
cp -r skill ~/.claude/skills/git-context
```

Claude will then use `git ctx` commands for:
- Saving decisions and context
- Planning with markdown checklists
- Coordinating with other agents via tasks and locks

See [`skill/SKILL.md`](skill/SKILL.md) for the full skill definition.

## Planning Pattern (Manus-style)

Use memory entries with checkboxes for structured planning:

```bash
git ctx add --title "Plan: Feature X" << 'EOF'
## Goal
Implement feature X

## Phases
- [ ] Phase 1: Setup
- [ ] Phase 2: Core implementation
- [ ] Phase 3: Tests

## Status
**Currently in Phase 1**
EOF
```

Before major decisions, re-read the plan:
```bash
git ctx show <plan-id>
```

After each phase, update:
```bash
git ctx edit <plan-id>
```

## Implementation

The Go implementation lives at the repository root (`cmd/git-ctx`, `internal/`). Local storage is complete; shared storage via git refs is planned.

## Inspiration

- [git-bug](https://github.com/git-bug/git-bug) — Distributed bug tracker in git refs
- [planning-with-files](https://github.com/OthmanAdi/planning-with-files) — Manus-style planning patterns
- [cc-mirror](https://github.com/numman-ali/cc-mirror) — Multi-agent task coordination

## License

Apache License 2.0 — see [LICENSE](LICENSE)
