## Context

The melon CLI is currently distributed as a single binary named `mln`, built from `cmd/mln`. All cobra commands are defined directly in that package. To support a `melon` binary with identical behavior, we need a second entrypoint that shares the same command tree without duplicating code.

The binary is distributed two ways:
1. Via GoReleaser GitHub release archives (direct download / `go install`)
2. Via the npm package `@playsthisgame/melon` (postinstall downloads the release archive and extracts `mln`)

Both distribution channels must be updated to expose `melon` alongside `mln`.

## Goals / Non-Goals

**Goals:**
- `melon <subcommand>` works identically to `mln <subcommand>` for all commands and flags
- Both binaries are present in every release archive
- npm global install registers both `mln` and `melon` bin entries
- `melon --version` and `melon --help` display correctly (showing `melon` not `mln`)
- No duplication of CLI logic

**Non-Goals:**
- Deprecating or removing `mln`
- Any behavioral difference between the two binaries
- Separate versioning for the `melon` binary

## Decisions

### 1. Extract CLI into `internal/cli`, add a thin `cmd/melon` entrypoint

The cobra root command and all subcommand wiring currently live in `cmd/mln`. This logic is extracted to `internal/cli` (a `Run(name, version string)` function that accepts the binary name so help text renders correctly). Both `cmd/mln/main.go` and `cmd/melon/main.go` become one-liners that call `cli.Run("mln", version)` and `cli.Run("melon", version)` respectively.

**Alternatives considered:**
- *Symlink at install time*: Fragile across platforms and package managers. Doesn't work cleanly on Windows.
- *Single binary that detects `os.Args[0]`*: Works, but produces one binary that must be renamed/symlinked by the installer — same fragility, more magic.
- *`cmd/melon` that just `exec`s `mln`*: Introduces a subprocess dependency; breaks if only one binary is on PATH.

The extract-to-internal approach is clean, testable, and idiomatic Go.

### 2. GoReleaser builds two binaries in a shared archive

A second `builds` entry in `.goreleaser.yaml` produces the `melon` binary from `./cmd/melon`. Both binaries are included in the same archive (`mln_<version>_<os>_<arch>.tar.gz`) so the postinstall script only needs one download.

**Alternatives considered:**
- *Separate archive for `melon`*: Doubles release asset count and complicates the postinstall download logic.
- *Rename `mln` to `melon` and ship a `mln` shim*: Would be a breaking change for existing users.

### 3. npm: add `"melon"` bin entry, postinstall creates a copy/symlink

`package.json` gains `"melon": "./bin/mln.js"` — the existing wrapper script works for both names since it just launches the downloaded binary. The postinstall script extracts both `mln` and `melon` from the archive and copies them to the package `bin/` directory.

On non-Windows the postinstall can create a hard link or copy; on Windows a copy is used (symlinks require elevated privileges).

## Risks / Trade-offs

- **Archive size increases slightly** — two near-identical Go binaries per archive. Acceptable given binary size (~10 MB each).
- **`go install` installs only one name** — `go install github.com/playsthisgame/melon/cmd/mln@latest` installs `mln`; users must separately run `go install github.com/playsthisgame/melon/cmd/melon@latest` for the `melon` alias. This is a known limitation of `go install` and is acceptable.
- **Refactor risk** — moving code from `cmd/mln` to `internal/cli` must not break existing behavior. Mitigated by the existing test suite.

## Open Questions

- None — implementation path is clear.
