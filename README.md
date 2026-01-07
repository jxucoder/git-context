# git-context

**Distributed, offline-first context storage embedded in git, with multi-agent support.**

Store coding context, decisions, and tasks directly in your git repository. Works offline. Syncs with push/pull. Designed for AI-assisted development and multi-agent collaboration.

## Features

- **Embedded in git**: Context lives in `.git/` — no external database, no server
- **Offline-first**: Works without network, syncs when ready
- **Distributed**: Every clone has full context history
- **Multi-agent ready**: Tasks, locks, and coordination primitives for AI agents
- **Two storage modes**: Local (private) and Shared (syncs with team)
- **Markdown-first**: Human-readable context entries

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
curl -L -o git-ctx https://github.com/jxucoder/git-context/releases/download/v0.2.0/git-ctx-darwin-arm64
chmod +x git-ctx
mv git-ctx ~/.local/bin/
git config --global alias.ctx '!git-ctx'
```

### Build from Source

```bash
cd go
make install
```

## Quick Start

```bash
# Add context (local by default)
git ctx add --title "Why JWT for auth"
# Opens editor, or use: git ctx add -t "Title" -m "Content"

# List entries
git ctx list

# Show an entry
git ctx show abc123

# Search
git ctx search "auth"

# Share with team (use --shared flag)
git ctx add --shared --title "Team coding standards"
git ctx push
```

## Tasks (Multi-Agent)

```bash
# Create tasks
git ctx task add "Implement user auth"
git ctx task add --shared "Team task"   # Syncs with push/pull

# List and claim
git ctx task list
git ctx task claim task-abc123

# Complete
git ctx task done task-abc123

# Add comments
git ctx task comment task-abc123 "Using bcrypt for passwords"
```

## Storage Model

Context is stored inside `.git/`, keeping your working directory clean.

| Mode | Location | Syncs? | Use Case |
|------|----------|--------|----------|
| Local (default) | `.git/context/` | No | Personal notes, drafts |
| Shared | `.git/refs/context/` | Yes | Team decisions, coordination |

Use `--shared` flag to store in shared storage. Sync with `git ctx push/pull`.

## Commands

### Memory (Context Entries)

| Command | Description |
|---------|-------------|
| `git ctx add [--title "T"] [-m "content"]` | Add entry |
| `git ctx list [--all]` | List entries |
| `git ctx show <id>` | View entry |
| `git ctx edit <id>` | Edit entry |
| `git ctx rm <id>` | Remove entry |
| `git ctx search "query"` | Search entries |

### Tasks

| Command | Description |
|---------|-------------|
| `git ctx task add "title"` | Create task |
| `git ctx task list [--all]` | List tasks |
| `git ctx task show <id>` | View task details |
| `git ctx task claim <id>` | Take ownership |
| `git ctx task done <id>` | Mark complete |
| `git ctx task comment <id> "msg"` | Add comment |

### Sync

| Command | Description |
|---------|-------------|
| `git ctx push` | Push shared entries to remote |
| `git ctx pull` | Pull shared entries from remote |

### Flags

| Flag | Description |
|------|-------------|
| `--shared`, `-s` | Use shared storage |
| `--all`, `-a` | Show both local and shared |
| `--json` | Output as JSON |

## Multi-Agent Workflow

1. **Team lead creates shared tasks:**
   ```bash
   git ctx task add --shared "Implement auth"
   git ctx task add --shared "Setup database"
   git ctx push
   ```

2. **Agents pull and claim work:**
   ```bash
   git ctx pull
   git ctx task list --shared
   git ctx task claim task-abc123
   git ctx push
   ```

3. **Agents complete and release:**
   ```bash
   git ctx task done task-abc123
   git ctx push
   ```

First to push wins. Conflicts are avoided through claiming.

## Use with Claude (AI Skill)

Copy the skill to your Claude skills directory:

```bash
cp -r skill ~/.claude/skills/git-context
```

Claude will then use `git ctx` commands for:
- Saving decisions and context
- Planning with markdown checklists
- Coordinating with other agents via tasks

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

The `go/` directory contains the Go implementation with full local storage support and sync capabilities.

## Inspiration

- [git-bug](https://github.com/git-bug/git-bug) — Distributed bug tracker in git refs
- [planning-with-files](https://github.com/OthmanAdi/planning-with-files) — Manus-style planning patterns
- [cc-mirror](https://github.com/numman-ali/cc-mirror) — Multi-agent task coordination

## License

Apache License 2.0 — see [LICENSE](LICENSE)
