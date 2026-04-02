## Context

Three artifacts are currently named with the short `mln` prefix:
- `.mln/` — the local package cache directory inside a project
- `mln.yaml` — the project manifest
- `mln.lock` — the reproducible lock file

The tool's full name is `melon`. Using the full name for on-disk artifacts improves discoverability. The single constant that drives the store path (`StoreDir = ".mln"`) and the two hardcoded strings in the CLI commands (`"mln.yaml"` / `"mln.lock"`) are the only things that need to change. No package interfaces, types, or algorithms are affected.

## Goals / Non-Goals

**Goals:**
- Change `StoreDir` constant in `internal/store/store.go` from `.mln` to `.melon`
- Change manifest filename from `mln.yaml` to `melon.yml` in all CLI commands and tests
- Change lockfile filename from `mln.lock` to `melon.lock` in all CLI commands and tests
- Update all comments, help strings, and user-facing messages referencing the old names
- Update `.gitignore`

**Non-Goals:**
- No changes to package public APIs (`manifest.Load`, `lockfile.Load`, etc.) — they accept arbitrary paths
- No migration tooling for existing `mln.yaml` / `.mln/` projects (out of scope for MVP)
- No binary rename — `mln` stays as the CLI command name

## Decisions

**Single constant for the store dir (`StoreDir`).**
All store path construction already flows through `store.StoreDir`. Changing one constant covers all callers — no search-and-replace needed for the store path.

**Manifest and lock filenames are hardcoded at the CLI layer.**
`install_cmd.go` and `init_cmd.go` construct these paths with `filepath.Join(dir, "mln.yaml")` etc. The simplest fix is replacing those string literals directly. No need to introduce new constants.

**File extension change: `.yaml` → `.yml`.**
The user requested `melon.yml` (not `melon.yaml`). Both are valid YAML extensions; `.yml` is more common in tooling. The parser (`gopkg.in/yaml.v3`) is extension-agnostic.

## Risks / Trade-offs

- **Breaking change for existing users** → Any project with `mln.yaml` / `.mln/` will need to rename their files. Acceptable for pre-1.0 / no established user base yet.
- **Test string churn** → Many test files hardcode `"mln.yaml"` / `"mln.lock"` / `".mln"`. All must be updated; missing one will cause test failures that are easy to catch.
