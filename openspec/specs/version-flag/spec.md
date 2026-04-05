### Requirement: Version flag prints CLI version and exits
`mln` SHALL support `--version` and `-v` flags that print the current CLI version to stdout and exit with code 0.

#### Scenario: --version flag
- **WHEN** the user runs `mln --version`
- **THEN** the output contains the CLI version (e.g. `mln version 0.1.3`) and the process exits with code 0

#### Scenario: -v shorthand flag
- **WHEN** the user runs `mln -v`
- **THEN** the output is identical to `mln --version` and the process exits with code 0

#### Scenario: Dev build version
- **WHEN** the binary was built without ldflags version injection
- **THEN** `mln --version` outputs `mln version dev`

### Requirement: Version is injected at build time
The CLI version SHALL be set via `-ldflags "-X main.version=<value>"` during the GoReleaser build so that release binaries carry the correct git tag version.

#### Scenario: Release binary version matches git tag
- **WHEN** GoReleaser builds the binary for tag `v0.2.0`
- **THEN** `mln --version` outputs `mln version 0.2.0`
