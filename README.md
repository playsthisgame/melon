<p align="center">
  <img src="assets/melon-logo.png" alt="melon" width="300" />
</p>

<h1 align="center">melon</h1>

<p align="center">
  A dependency manager for agentic markdown — skills, agents, and prompts for your AI tools.
</p>

<p align="center">
  <a href="#installation">Installation</a> ·
  <a href="#quick-start">Quick Start</a> ·
  <a href="#how-it-works">How it works</a> ·
  <a href="#manifest-reference">Manifest Reference</a> ·
  <a href="#commands">Commands</a>
</p>

---

## What is melon?

Melon manages markdown-based packages — skills, agents, workflows, personas, and memory files — that AI coding assistants read as context. It resolves dependencies from GitHub, fetches them into a local cache, and places them into your agent's expected directory (e.g. `.claude/skills/`) so they are available immediately.

## Installation

```sh
go install github.com/playsthisgame/melon/cmd/mln@latest
```

Requires Git to be available on your `PATH`.

## Quick Start

**1. Initialize a project**

```sh
mln init
```

This creates a `melon.yml` manifest and the `.melon/` cache directory. You'll be prompted for a package name, type, and which AI agents you target.

**2. Add a dependency**

Edit `melon.yml` directly:

```yaml
dependencies:
  github.com/anthropics/skills/skills/skill-creator: "main"
  github.com/alice/pdf-skill: "^1.2.0"
```

Dependency names are GitHub paths. You can use:
- A full repo — `github.com/owner/repo`
- A monorepo subdirectory — `github.com/owner/repo/path/to/skill`
- A GitHub web URL — `github.com/owner/repo/tree/main/path/to/skill` (the `tree/<branch>/` is stripped automatically)

Versions can be a semver constraint (`^1.2.0`, `~2.0.0`, `1.0.0`) or a branch name (`"main"`).

**3. Install**

```sh
mln install
```

Melon resolves each dependency, fetches it via sparse git checkout, writes `melon.lock`, and places skills into your agent directories.

```
  resolving github.com/alice/pdf-skill (^1.2.0)...
  fetching github.com/alice/pdf-skill@1.2.0...
  + github.com/alice/pdf-skill@1.2.0
  linked github.com/alice/pdf-skill -> .claude/skills/pdf-skill
```

## How it works

```
melon.yml          — declares your dependencies and target agents      ← commit
melon.lock         — pins exact versions, git tags, and content hashes ← commit
.melon/            — local cache; one directory per dep@version        ← commit
.claude/skills/    — symlinks into .melon/ created by mln install      ← commit
```

Skills are fetched once into `.melon/` and symlinked into agent directories. Committing everything means skills are available to the whole team immediately after cloning — no extra step needed. Re-running `mln install` is idempotent: it skips fetches whose tree hash already matches and recreates symlinks in place.

## Manifest Reference

```yaml
# melon.yml

name: my-agent
version: 0.1.0

# type: skill | agent | workflow | persona | memory
type: agent

description: "My coding agent"

dependencies:
  github.com/anthropics/skills/skills/skill-creator: "main"
  github.com/alice/pdf-skill: "^1.2.0"

# agent_compat drives where mln install places skills.
# Melon knows the conventions for each agent automatically:
#   claude-code    -> .claude/skills/
#   cursor         -> .agents/skills/
#   windsurf       -> .windsurf/skills/
#   roo            -> .roo/skills/
#   ... (and more)
agent_compat:
  - claude-code

# outputs is optional. Use it to override the automatic placement paths.
# outputs:
#   .claude/skills/: "*"

tags: []
```

### Supported package types

| Type | Description |
|---|---|
| `skill` | A reusable skill or tool instructions for an AI agent |
| `agent` | A complete agent definition with its own context and skills |
| `workflow` | A multi-step process or automation definition |
| `persona` | A personality or role definition |
| `memory` | Persistent knowledge or context files |

### Supported agents

| Agent | Project skills directory |
|---|---|
| `claude-code` | `.claude/skills/` |
| `cursor` | `.agents/skills/` |
| `windsurf` | `.windsurf/skills/` |
| `roo` | `.roo/skills/` |
| `codex` | `.agents/skills/` |
| `opencode` | `.agents/skills/` |
| `gemini-cli` | `.agents/skills/` |
| `github-copilot` | `.agents/skills/` |
| `cline` | `.agents/skills/` |
| `amp` | `.agents/skills/` |

## Commands

### `mln init`

Scaffold a new `melon.yml` and create the `.melon/` store directory.

```sh
mln init
mln init --yes        # accept all defaults (for scripting)
mln init --dir ./app  # initialize in a different directory
```

### `mln install`

Resolve dependencies, fetch them into `.melon/`, write `melon.lock`, and symlink skills into agent directories.

```sh
mln install
mln install --frozen    # fail if melon.lock would change (useful in CI)
mln install --no-place  # fetch and lock only — skip placement into agent dirs
```

### `mln add`

*(Coming soon)* Add a dependency to `melon.yml` and run install.

### `mln remove`

*(Coming soon)* Remove a dependency from `melon.yml` and clean up.

## Lock file

`melon.lock` is generated by `mln install` and should be committed to version control. It pins the exact version, git tag, repo URL, subdirectory, and a SHA-256 tree hash for each dependency.

```yaml
generated_at: "2025-03-31T12:00:00Z"
dependencies:
  - name: github.com/alice/pdf-skill
    version: "1.2.0"
    git_tag: v1.2.0
    repo_url: https://github.com/alice/pdf-skill
    subdir: ""
    entrypoint: SKILL.md
    tree_hash: "sha256:abc123..."
    files:
      - SKILL.md
```

Use `--frozen` in CI to ensure the lock file is always up to date:

```sh
mln install --frozen
```

## License

[MIT](LICENSE)
