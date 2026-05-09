## Requirements

### Requirement: JSON output mode is opt-in via a per-command flag

Commands that support machine-readable output SHALL accept a `--json` flag. When `--json` is set, the command SHALL write a single valid JSON document to stdout and suppress all TUI output (spinners, progress bars, styled text).

#### Scenario: JSON flag produces valid JSON on stdout

- **WHEN** a supported command is run with `--json`
- **THEN** stdout contains exactly one valid JSON document and nothing else

#### Scenario: TUI is suppressed in JSON mode

- **WHEN** stdout is a terminal and `--json` is set
- **THEN** no spinners, progress bars, or lipgloss-styled output appear on stdout

### Requirement: Errors are written as JSON to stderr in JSON mode

When `--json` is set and the command encounters an error, the error SHALL be written to stderr as `{"error": "<message>"}` and the command SHALL exit with a non-zero code.

#### Scenario: Error in JSON mode

- **WHEN** a supported command is run with `--json` and the command fails
- **THEN** stderr contains `{"error": "<message>"}` and stdout is empty
