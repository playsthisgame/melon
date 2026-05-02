## Context

Currently `melon init` creates `.melon/` and writes `melon.yaml` but never touches `.gitignore`. This means managed symlinks (e.g. `.claude/skills/golang-pro`) and the `.melon/` cache are silently tracked by git unless the user manually adds them to `.gitignore`. The project treats `.melon/` as a transient cache (analogous to `node_modules/`), but gives users no automated way to opt into that model.

The manifest schema (`internal/manifest/schema.go`) is a plain struct — adding a `Vendor bool` field with a pointer-or-default pattern is straightforward. The gitignore file is a plain text file with one pattern per line; no external library is needed.

## Goals / Non-Goals

**Goals:**
- Add `vendor` boolean to `melon.yaml` schema, default `true` (preserves current behavior)
- `melon init` prompts for vendoring preference; writes `vendor: false` only when the user opts out
- When `vendor: false`, `melon install` syncs `.gitignore` with `.melon/` and all managed symlink paths
- When `vendor: false`, `melon add` appends new symlink paths to `.gitignore`
- When `vendor: false`, `melon remove` removes stale symlink paths from `.gitignore`
- Create `.gitignore` if it does not exist when `vendor: false` and entries need to be written

**Non-Goals:**
- Removing entries from `.gitignore` that the user added manually
- Supporting `.gitignore` files in subdirectories or parent directories
- Any UI for switching `vendor` after init (user edits `melon.yaml` directly)
- Handling the case where symlinks live in a non-standard location (i.e. `outputs` override — best-effort only)

## Decisions

### 1. Default `vendor: true`

**Decision:** Vendor mode is opt-out, not opt-in.

**Rationale:** Changing the default to `false` would be a breaking change for existing users who are already committing `.melon/` and their symlinks. Opt-out preserves the current behavior for all existing `melon.yaml` files that lack the field.

**Alternative considered:** Default `false` (non-vendoring). Rejected because it would silently start modifying `.gitignore` for users who upgrade melon without reading release notes.

### 2. New `internal/gitignore` package

**Decision:** Encapsulate all `.gitignore` read/write logic in a dedicated package with pure functions: `EnsureEntries`, `RemoveEntries`, and `ContainsEntry`.

**Rationale:** Keeps gitignore manipulation testable in isolation. The CLI commands stay thin — they compute which paths need to be added/removed and delegate to this package. No external dependency needed; `.gitignore` is line-oriented text.

**Alternative considered:** Inline the logic in each command. Rejected because three commands need it; a shared package avoids duplication and makes the logic easier to test.

### 3. Idempotent `EnsureEntries`

**Decision:** `EnsureEntries` checks whether each entry already exists before appending. It appends a melon-managed block comment header on first write so the section is clearly labelled in the file.

**Rationale:** `melon install` is designed to be re-runnable safely. The gitignore sync must follow the same contract.

### 4. Symlink paths computed from lock file

**Decision:** During `install`, the set of paths to gitignore is derived from the lock file (post-resolve), not from the manifest. During `add`/`remove`, only the affected dep's paths are touched.

**Rationale:** The lock file is the ground truth for what is actually placed on disk. Using it avoids drift between what is installed and what is gitignored.

### 5. `melon init` prompt placement

**Decision:** Add the vendoring question as the last prompt in the init flow, after tool compat selection, with a clear default of `y` (vendor = true / keep in git).

**Rationale:** Most users running `melon init` in a new project will not have a strong opinion. Defaulting to yes avoids surprising them with an automatically modified `.gitignore`. Users who know they want non-vendoring can opt out explicitly.

## Risks / Trade-offs

- **`outputs` override paths** — When `melon.yaml` uses `outputs` instead of `tool_compat`, the symlink paths are user-defined. The gitignore sync will still work because symlink paths are derived from the placer's output, but this path is less tested. → Mitigation: cover with an integration test using a custom `outputs` block.

- **Existing `.gitignore` formatting** — The package appends entries without reformatting the file. If the user has trailing whitespace or non-unix line endings the file could look inconsistent. → Mitigation: normalize to `\n` on write; document that melon appends entries at the end.

- **`vendor: true` files already in git** — Switching from `vendor: true` to `vendor: false` on an existing project does not automatically run `git rm --cached`. Melon cannot safely do this. → Mitigation: print a hint after the first `melon install` with `vendor: false` telling the user to run `git rm --cached` for the newly-ignored paths.

## Migration Plan

No migration required. Existing `melon.yaml` files without a `vendor` field parse as `vendor: true` (zero value for `*bool` treated as true, or use explicit pointer with nil → true). No `.gitignore` changes occur for existing projects unless the user adds `vendor: false` to their manifest.
