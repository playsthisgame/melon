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
  <a href="#commands">Commands</a>
</p>

<p align="center">
  <a href="https://github.com/playsthisgame/melon/actions/workflows/release.yml"><img src="https://github.com/playsthisgame/melon/actions/workflows/release.yml/badge.svg" alt="Release" /></a>
  <a href="https://www.npmjs.com/package/@playsthisgame/melon"><img src="https://img.shields.io/npm/v/@playsthisgame/melon" alt="npm version" /></a>
  <a href="https://www.npmjs.com/package/@playsthisgame/melon"><img src="https://img.shields.io/npm/dt/@playsthisgame/melon" alt="total downloads" /></a>
  <a href="https://pkg.go.dev/github.com/playsthisgame/melon"><img src="https://pkg.go.dev/badge/github.com/playsthisgame/melon.svg" alt="Go Reference" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/playsthisgame/melon" alt="License" /></a>
</p>

---

## What is melon?

Melon manages markdown-based packages that AI coding assistants read as context. It resolves dependencies from GitHub, fetches them into a local cache, and places them into your agent's expected directory (e.g. `.claude/skills/`) so they are available immediately.

## See it in action

![melon demo](https://raw.githubusercontent.com/playsthisgame/melon/main/assets/demo.gif)

## Installation

**📦 Global install**

```sh
npm install -g @playsthisgame/melon
```

**🐹 Go**

```sh
go install github.com/playsthisgame/melon/cmd/melon@latest
```

Requires Git to be available on your `PATH`.

## Quick Start

**1. Initialize a project**

```sh
melon init
```

This creates a `melon.yaml` manifest and the `.melon/` cache directory. You'll be prompted for a package name, type, and which AI tools you use.

**2. Add a dependency**

Edit `melon.yaml` directly:

```yaml
dependencies:
  github.com/playsthisgame/melon-index/skills/video-to-gif: "^1.0.2"
  github.com/anthropics/skills/skills/skill-creator: "main"
```

Dependency names are GitHub paths. You can use:
- A full repo: `github.com/owner/repo`
- A monorepo subdirectory: `github.com/owner/repo/path/to/skill`
- A GitHub web URL: `github.com/owner/repo/tree/main/path/to/skill` (the `tree/<branch>/` is stripped automatically)

Versions can be a semver constraint (`^1.2.0`, `~2.0.0`, `1.0.0`) or a branch name (`"main"`).

**3. Install**

```sh
melon install
```

Melon resolves each dependency, fetches it via sparse git checkout, writes `melon.lock`, and places skills into your tool directories.

```
  resolving github.com/playsthisgame/melon-index/skills/video-to-gif (^1.0.2)...
  fetching github.com/playsthisgame/melon-index/skills/video-to-gif@1.0.2...
  + github.com/playsthisgame/melon-index/skills/video-to-gif@1.0.2
  linked github.com/playsthisgame/melon-index/skills/video-to-gif -> .claude/playsthisgame/video-to-gif
```

## How it works

```
melon.yaml          — declares your dependencies and target AI tools    ← commit
melon.lock         — pins exact versions, git tags, and content hashes ← commit
.melon/            — local cache; one directory per dep@version
.claude/skills/    — symlinks into .melon/ created by melon install
```

Skills are fetched once into `.melon/` and symlinked into the configured tools directories.

### Vendoring vs. gitignore management

By default (`vendor: true`, the implicit setting), `.melon/` and the generated symlinks are committed to your repo — anyone who clones gets the skills without running `melon install`. If you prefer to keep deps out of git, set `vendor: false` in `melon.yaml` (or opt out during `melon init`). Melon will automatically maintain a `.gitignore` block for `.melon/` and all managed symlink paths, keeping it in sync as you add or remove dependencies.

## Manifest Reference

```yaml
# melon.yaml

name: my-agent
version: 0.1.0

description: "My coding agent"

dependencies:
  github.com/anthropics/skills/skills/skill-creator: "main"
  github.com/playsthisgame/melon-index/skills/video-to-gif: "^1.0.2"

# tool_compat drives where melon install places skills.
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

# vendor controls whether melon manages .gitignore for its cache and symlinks.
# true (default): .melon/ and skill symlinks are committed to your repo.
# false: melon auto-manages .gitignore — add/remove keeps it in sync.
# vendor: false

# index is optional. Use it to point melon search/info at private registries.
# Melon supports GitHub paths and web URLs:
#   - github.com/owner/repo/path/to/index.yaml
#   - github.com/owner/repo/tree/main/path/to/index.yaml
#   - https://raw.githubusercontent.com/owner/repo/main/path/to/index.yaml
# index:
#   urls:
#     - github.com/my-company/skill-registry/index.yaml
#     - https://example.com/index.yaml
#   public_index: true  # false = suppress the default public melon index, default is true

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

### `melon init`

Scaffold a new `melon.yaml` and create the `.melon/` store directory.

```sh
melon init
melon init --yes        # accept all defaults (for scripting)
melon init --dir ./app  # initialize in a different directory
```

### `melon install`

Resolve dependencies, fetch them into `.melon/`, write `melon.lock`, and symlink skills into tool directories.

```sh
melon install
melon install --frozen    # fail if melon.lock would change (useful in CI)
melon install --no-place  # fetch and lock only — skip placement into agent dirs
```

### `melon add`

Add a dependency to `melon.yaml` and run install. If no version is specified, the latest semver tag is resolved automatically.

```sh
melon add github.com/playsthisgame/melon-index/skills/video-to-gif          # resolves latest tag → ^1.2.0
melon add github.com/playsthisgame/melon-index/skills/video-to-gif@^1.0.0   # explicit constraint
melon add github.com/playsthisgame/melon-index/skills/video-to-gif@main     # branch pin
```

### `melon remove`

Remove a dependency from `melon.yaml`, unlink its agent symlinks, and delete its `.melon/` cache entry.

```sh
melon remove github.com/playsthisgame/melon-index/skills/video-to-gif
```

### `melon update`

Update dependencies to the latest version satisfying their existing constraints. Without arguments, shows an interactive multi-select list (TTY only) — the first option updates all deps. With a dep name, targets that dep directly.

```sh
melon update                                                    # interactive: pick which deps to update
melon update github.com/playsthisgame/melon-index/skills/video-to-gif  # update a specific dep
```

If a newer major version exists outside your constraint, a hint is printed showing how to upgrade with `melon add`.

### `melon outdated`

Show which dependencies have newer versions available, without modifying anything. Exits with code 1 if any dep is outdated — useful in CI.

```sh
melon outdated
```

Output includes current locked version, latest compatible version, and absolute latest (when a newer major exists outside your constraint). Branch-pinned deps are skipped.

### `melon list`

Show installed skills and audit installation state.

```sh
melon list                # list all installed skills
melon list --pending      # show skills in melon.yaml not yet installed
melon list --check        # verify skill symlinks exist in all tool directories
melon list --pending --check  # combine both checks
```

Default output:

```text
  github.com/anthropics/skills/skills/skill-creator  main
  github.com/playsthisgame/melon-index/skills/video-to-gif   1.2.0
```

With `--check`:

```text
  OK       github.com/playsthisgame/melon-index/skills/video-to-gif@1.2.0   .claude/skills/video-to-gif
  MISSING  github.com/anthropics/skills/skills/skill-creator@main    .claude/skills/skill-creator
```

With `--pending`:

```text
Pending (not installed):
  github.com/owner/repo/some-new-skill
```

### `melon search`

Search for skills by keyword against the [melon-index](https://github.com/playsthisgame/melon-index) curated list. In a terminal, results are shown in an interactive list — navigate with `↑↓`, press `Enter` to select, and melon will offer to run `melon add` for you.

```sh
melon search spec          # find spec-related skills
melon search git workflow  # find git workflow skills
```

Featured skills appear at the top of results. If nothing matches, melon will tell you and suggest submitting to the index.

If your project has an `index` block in `melon.yaml`, search queries your custom registry instead of (or in addition to) the public melon index. Set `public_index: true` to see only your private skills.

### `melon info`

Show metadata for a specific skill — description, author, and available versions — before installing it.

```sh
melon info github.com/playsthisgame/melon-index/skills/video-to-gif
melon info github.com/owner/repo/path/to/skill
```

### `melon clean`

Remove cached entries in `.melon/` that are no longer referenced by `melon.lock`. Any corresponding symlinks in agent skill directories are removed as well. This reclaims disk space left behind after removing or upgrading dependencies.

```sh
melon clean
```

If `melon.lock` does not exist, the command prints a reminder to run `melon install` first and exits without modifying anything.

## Lock file

`melon.lock` is generated by `melon install` and should be committed to version control. It pins the exact version, git tag, repo URL, subdirectory, and a SHA-256 tree hash for each dependency.

```yaml
generated_at: "2025-03-31T12:00:00Z"
dependencies:
  - name: github.com/playsthisgame/melon-index/skills/video-to-gif
    version: "1.2.0"
    git_tag: v1.2.0
    repo_url: https://github.com/playsthisgame/melon-index
    subdir: "skills/video-to-gif"
    entrypoint: SKILL.md
    tree_hash: "sha256:abc123..."
    files:
      - SKILL.md
```

Use `--frozen` in CI to ensure the lock file is always up to date:

```sh
melon install --frozen
```

## Why melon?

As AI coding assistants become more capable, teams are building and sharing libraries of skills. Without a proper dependency manager, keeping these skills consistent across developers, environments, and CI becomes a manual, error-prone process.

**Melon gives you a single source of truth.** Define all the skills your project needs in one `melon.yaml` file, commit it alongside your code, and every developer (and your CI pipeline) gets exactly the same set of skills with a single `melon install`.

**Skills are versioned, not just copied.** Melon pins exact versions, git tags, and SHA-256 content hashes in `melon.lock`. If a skill author publishes a breaking change, your team won't silently pick it up, you'll see the diff in the lock file and upgrade intentionally. This means you can trust that the skill running in CI today is the same one that ran last week.

**It works naturally with CI.** Run `melon install --frozen` in your pipeline to fail fast if the lock file is out of sync with the manifest. No surprises, no drift. With the default `vendor: true` setting, `.melon/` and the generated symlinks are committed to the repo so CI doesn't even need network access — everything is already there. If you prefer to keep deps out of git, set `vendor: false` and CI will fetch them fresh on each run using the pinned versions in `melon.lock`.

**Works across your whole team and all your tools.** List the AI tools your project uses under `tool_compat` and melon places each skill into every agent's expected directory at once. One manifest, one install command, every agent ready to go.

## Why melon instead of npx skill installers?

Many agent skill collections ship a one-liner like `npx install-skill <name>` that copies files into your project. Melon takes a different approach:

| | melon | npx installers |
| --- | --- | --- |
| **Reproducibility** | `melon.lock` pins exact versions and content hashes | Each run may fetch a different version |
| **Transitive deps** | Resolves the full dependency graph | Usually single-package only |
| **Multiple agents** | `tool_compat` places skills for all your tools at once | Typically one target agent |
| **Offline / CI** | Already-fetched deps are cached in `.melon/` | Always fetches from the network |
| **Node.js required** | No — pure Go binary, no runtime needed | Yes |
| **Removal** | `melon remove` unlinks symlinks and purges the cache | Usually manual |

## License

[MIT](LICENSE)
