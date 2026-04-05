### Requirement: The melon binary is a full alias for mln
The `melon` binary SHALL be distributed alongside `mln` and SHALL support all the same subcommands (`init`, `install`, `add`, `remove`), flags, and behaviors. Invoking `melon <subcommand>` SHALL produce identical results to `mln <subcommand>`.

#### Scenario: melon subcommands work identically to mln
- **WHEN** a user runs `melon install`, `melon add`, `melon remove`, or `melon init`
- **THEN** the command SHALL execute with the same behavior as the corresponding `mln` command

#### Scenario: melon --version displays the installed version
- **WHEN** a user runs `melon --version`
- **THEN** the output SHALL display the current melon version

#### Scenario: melon --help displays help text with melon as the binary name
- **WHEN** a user runs `melon --help`
- **THEN** the help text SHALL reference `melon` as the binary name (not `mln`)

### Requirement: Both mln and melon are installed by npm global install
After `npm install -g @playsthisgame/melon`, both `mln` and `melon` SHALL be available as executable commands on the user's PATH.

#### Scenario: npm global install registers melon bin
- **WHEN** a user runs `npm install -g @playsthisgame/melon`
- **THEN** both `mln --version` and `melon --version` SHALL succeed

### Requirement: Both mln and melon are available via go install
The repository SHALL expose a `cmd/melon` package so users can install the `melon` binary via `go install github.com/playsthisgame/melon/cmd/melon@latest`.

#### Scenario: go install cmd/melon installs the melon binary
- **WHEN** a user runs `go install github.com/playsthisgame/melon/cmd/melon@latest`
- **THEN** the `melon` binary SHALL be placed in `$GOPATH/bin` and SHALL be executable
