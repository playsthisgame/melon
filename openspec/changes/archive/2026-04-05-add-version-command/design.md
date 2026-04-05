## Context

`mln` is built with Cobra and distributed as a Go binary via GoReleaser. Currently there is no `--version` flag and no package-level version variable. The `mln init` command scaffolds `melon.yml` with a hardcoded `version: 0.1.0` string. GoReleaser does not currently inject the build version via ldflags.

## Goals / Non-Goals

**Goals:**
- `mln --version` / `mln -v` prints the CLI version and exits
- `mln init` writes the actual CLI version into the `version` field of the generated `melon.yml`
- Version is injected at build time by GoReleaser so release binaries always carry the correct value
- Dev builds (no ldflags) fall back to `"dev"`

**Non-Goals:**
- Structured/JSON version output
- Separate `mln version` subcommand (flag is sufficient)
- Changing the semver format of `melon.yml` project versions

## Decisions

**Use Cobra's built-in `Version` field rather than a custom flag**

Cobra's `rootCmd.Version = version` automatically wires `--version` and `-v`, prints `mln version <value>`, and exits. This is less code than a manual flag and follows the standard Cobra pattern.

**Inject version via ldflags at build time**

A package-level `var version = "dev"` in `main.go` is overwritten by GoReleaser using:
```
-ldflags "-X main.version={{ .Version }}"
```
This is the idiomatic Go approach — no runtime file reads, no embedded files, zero overhead.

**Pass version into init via a package-level variable**

`init_cmd.go` already imports `main` implicitly (same package). The `version` var declared in `main.go` is visible to `init_cmd.go` without any refactoring. The hardcoded `"0.1.0"` string in the init template is replaced with a `fmt.Sprintf` or direct substitution using `version`.

## Risks / Trade-offs

- [Dev builds show `"dev"` version in `melon.yml`] → Acceptable; developers know they're on a dev build. Could add a git-describe fallback later if needed.
- [Cobra's `-v` shorthand conflicts if another flag uses `-v`] → Currently `--verbose` uses no shorthand, so no conflict.

## Open Questions

None — straightforward change with well-established Go patterns.
