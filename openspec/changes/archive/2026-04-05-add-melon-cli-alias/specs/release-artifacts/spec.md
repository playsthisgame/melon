## MODIFIED Requirements

### Requirement: A .goreleaser.yaml and GitHub Actions release workflow are present
The repository SHALL contain a `.goreleaser.yaml` that produces prebuilt binaries for both `mln` and `melon` for macOS arm64, macOS amd64, Linux amd64, and Windows amd64, and a `.github/workflows/release.yml` that triggers GoReleaser on git tag push (tags matching `v*`). Both binaries SHALL be included in the same release archive.

#### Scenario: Release workflow triggers on tag push
- **WHEN** a git tag matching `v*` is pushed
- **THEN** the GitHub Actions workflow SHALL trigger and run GoReleaser

#### Scenario: Release archives contain both mln and melon binaries
- **WHEN** GoReleaser produces a release archive (e.g., `mln_<version>_darwin_arm64.tar.gz`)
- **THEN** the archive SHALL contain both the `mln` and `melon` binaries (or `mln.exe` and `melon.exe` on Windows)
