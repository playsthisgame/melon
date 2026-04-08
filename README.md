<p align="center">
  <img src="https://raw.githubusercontent.com/playsthisgame/melon/main/assets/melon-logo.png" alt="melon" width="300" />
</p>

<h1 align="center">melon</h1>

<p align="center">
  A dependency manager for agentic markdown — skills, agents, and prompts for your AI tools.
</p>

<p align="center">
  <a href="#installation">Installation</a> ·
  <a href="#quick-start">Quick Start</a> ·
  <a href="#why-melon">Why melon?</a> ·
  <a href="#how-it-works">How it works</a> ·
  <a href="#manifest-reference">Manifest Reference</a> ·
  <a href="#commands">Commands</a> ·
  <a href="#discovering-skills">Discovering Skills</a> ·
  <a href="#publishing-a-skill">Publishing a Skill</a>
</p>

<p align="center">
  <a href="https://github.com/playsthisgame/melon/actions/workflows/release.yml"><img src="https://github.com/playsthisgame/melon/actions/workflows/release.yml/badge.svg" alt="Release" /></a>
  <a href="https://www.npmjs.com/package/@playsthisgame/melon"><img src="https://img.shields.io/npm/v/@playsthisgame/melon" alt="npm version" /></a>
  <a href="https://pkg.go.dev/github.com/playsthisgame/melon"><img src="https://pkg.go.dev/badge/github.com/playsthisgame/melon.svg" alt="Go Reference" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/playsthisgame/melon" alt="License" /></a>
</p>

---

## What is melon?

Melon manages markdown-based packages that AI coding assistants read as context. It resolves dependencies from GitHub, fetches them into a local cache, and places them into your agent's expected directory (e.g. `.claude/skills/`) so they are available immediately.

## Installation

**📦 Global install**

```sh
npm install -g @playsthisgame/melon
```

**🐹 Go**

```sh
go install github.com/playsthisgame/melon/cmd/mln@latest
```

Requires Git to be available on your `PATH`.

## Quick Start

**1. Initialize a project**

```sh
mln init
```

This creates a `melon.yml` manifest and the `.melon/` cache directory. You'll be prompted for a package name, type, and which AI tools you use.

**2. Add a dependency**

Edit `melon.yml` directly:

```yaml
dependencies:
  github.com/playsthisgame/skills/agentic-spec-dev: "^1.0.0"
  github.com/anthropics/skills/skills/skill-creator: "main"
```

Dependency names are GitHub paths. You can use:
- A full repo: `github.com/owner/repo`
- A monorepo subdirectory: `github.com/owner/repo/path/to/skill`
- A GitHub web URL: `github.com/owner/repo/tree/main/path/to/skill` (the `tree/<branch>/` is stripped automatically)

Versions can be a semver constraint (`^1.2.0`, `~2.0.0`, `1.0.0`) or a branch name (`"main"`).

**3. Install**

```sh
mln install
```

Melon resolves each dependency, fetches it via sparse git checkout, writes `melon.lock`, and places skills into your tool directories.

```
  resolving github.com/playsthisgame/skills/agentic-spec-dev (^1.0.0)...
  fetching github.com/playsthisgame/skills/agentic-spec-dev@1.0.0...
  + github.com/playsthisgame/skills/agentic-spec-dev@1.0.0
  linked github.com/playsthisgame/skills/agentic-spec-dev -> .claude/playsthisgame/agentic-spec-dev
```

## How it works

```
melon.yml          — declares your dependencies and target AI tools    ← commit
melon.lock         — pins exact versions, git tags, and content hashes ← commit
.melon/            — local cache; one directory per dep@version        ← commit
.claude/skills/    — symlinks into .melon/ created by mln install      ← commit
```

Skills are fetched once into `.melon/` and symlinked into the configured tools directories.

## Manifest Reference

```yaml
# melon.yml

name: my-agent
version: 0.1.0

description: "My coding agent"

dependencies:
  github.com/anthropics/skills/skills/skill-creator: "main"
  github.com/playsthisgame/skills/agentic-spec-dev: "^1.0.0"

# tool_compat drives where mln install places skills.
# Melon knows the conventions for each agent automatically:
#   claude-code    -> .claude/skills/
#   cursor         -> .agents/skills/
#   windsurf       -> .windsurf/skills/
#   roo            -> .roo/skills/
#   ... (and more)
tool_compat:
  - claude-code

# outputs is optional. Use it to override the automatic placement paths.
# outputs:
#   .claude/skills/: "*"

tags: []
```

### Supported AI tools

| AI tool | Project skills directory |
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

Resolve dependencies, fetch them into `.melon/`, write `melon.lock`, and symlink skills into tool directories.

```sh
mln install
mln install --frozen    # fail if melon.lock would change (useful in CI)
mln install --no-place  # fetch and lock only — skip placement into agent dirs
```

### `mln add`

Add a dependency to `melon.yml` and run install. If no version is specified, the latest semver tag is resolved automatically.

```sh
mln add github.com/playsthisgame/skills/agentic-spec-dev          # resolves latest tag → ^1.2.0
mln add github.com/playsthisgame/skills/agentic-spec-dev@^1.0.0   # explicit constraint
mln add github.com/playsthisgame/skills/agentic-spec-dev@main     # branch pin
```

### `mln remove`

Remove a dependency from `melon.yml`, unlink its agent symlinks, and delete its `.melon/` cache entry.

```sh
mln remove github.com/playsthisgame/skills/agentic-spec-dev
```

### `mln search`

Search for skills by keyword. Checks the [melon-index](https://github.com/playsthisgame/melon-index) curated list first, then falls back to GitHub Topics (`melon-skill`) if no curated results are found. In a terminal, results are shown in an interactive list — press Enter to select and be prompted to install.

```sh
mln search spec          # find spec-related skills
mln search git workflow  # find git workflow skills
```

### `mln info`

Show metadata for a specific skill — description, author, and available versions — before installing it.

```sh
mln info github.com/playsthisgame/skills/agentic-spec-dev
mln info github.com/owner/repo/path/to/skill
```

## Lock file

`melon.lock` is generated by `mln install` and should be committed to version control. It pins the exact version, git tag, repo URL, subdirectory, and a SHA-256 tree hash for each dependency.

```yaml
generated_at: "2025-03-31T12:00:00Z"
dependencies:
  - name: github.com/playsthisgame/skills/agentic-spec-dev
    version: "1.2.0"
    git_tag: v1.2.0
    repo_url: https://github.com/playsthisgame/skills/agentic-spec-dev
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

## Why melon?

As AI coding assistants become more capable, teams are building and sharing libraries of skills. Without a proper dependency manager, keeping these skills consistent across developers, environments, and CI becomes a manual, error-prone process.

**Melon gives you a single source of truth.** Define all the skills your project needs in one `melon.yml` file, commit it alongside your code, and every developer (and your CI pipeline) gets exactly the same set of skills with a single `mln install`.

**Skills are versioned, not just copied.** Melon pins exact versions, git tags, and SHA-256 content hashes in `melon.lock`. If a skill author publishes a breaking change, your team won't silently pick it up, you'll see the diff in the lock file and upgrade intentionally. This means you can trust that the skill running in CI today is the same one that ran last week.

**It works naturally with CI.** Run `mln install --frozen` in your pipeline to fail fast if the lock file is out of sync with the manifest. No surprises, no drift. Because `.melon/` and the generated symlinks are committed to the repo, CI doesn't even need network access to place skills, everything is already there.

**Works across your whole team and all your tools.** List the AI tools your project uses under `tool_compat` and melon places each skill into every agent's expected directory at once. One manifest, one install command, every agent ready to go.

## Discovering Skills

Use `mln search` to find skills by keyword:

```sh
mln search spec
mln search git workflow
```

Search checks the [melon-index](https://github.com/playsthisgame/melon-index) — a curated list of reviewed skills — first. If no curated results match, it falls back to GitHub Topics, showing any public repo tagged `melon-skill`. Results from the fallback are marked as community-tagged.

In a terminal, results appear as an interactive list. Navigate with `↑↓`, press `Enter` to select, and melon will offer to run `mln add` for you. Use `mln info` to inspect a skill before installing:

```sh
mln info github.com/owner/repo/path/to/skill
```

## Publishing a Skill

There are two ways to make your skill discoverable:

### Option 1 — Submit to the curated index (recommended)

Open a pull request on [melon-index](https://github.com/playsthisgame/melon-index) adding an entry to `index.yml`:

```yaml
skills:
  - name: github.com/your-username/your-skill
    description: "What your skill does"
    author: your-username
    tags: [relevant, keywords]
```

Curated skills appear first in `mln search` results.

### Option 2 — Tag your repo with `melon-skill` (immediate)

Add the `melon-skill` topic to your GitHub repo (repo Settings → About → Topics). Your skill becomes discoverable immediately as a fallback result in `mln search` without waiting for index review.

Tag your release with a semver version (e.g. `v1.0.0`) so users can pin to a specific version with `mln add`.

## Why melon instead of npx skill installers?

Many agent skill collections ship a one-liner like `npx install-skill <name>` that copies files into your project. Melon takes a different approach:

| | melon | npx installers |
| --- | --- | --- |
| **Reproducibility** | `melon.lock` pins exact versions and content hashes | Each run may fetch a different version |
| **Transitive deps** | Resolves the full dependency graph | Usually single-package only |
| **Multiple agents** | `tool_compat` places skills for all your tools at once | Typically one target agent |
| **Offline / CI** | Already-fetched deps are cached in `.melon/` | Always fetches from the network |
| **Node.js required** | No — pure Go binary, no runtime needed | Yes |
| **Removal** | `mln remove` unlinks symlinks and purges the cache | Usually manual |

## License

[MIT](LICENSE)
