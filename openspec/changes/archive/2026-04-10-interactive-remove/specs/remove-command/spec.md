## MODIFIED Requirements

### Requirement: mln remove removes a named dependency from melon.yml
When the user runs `mln remove <name>`, the command SHALL remove the entry for `<name>` from `melon.yaml` and write the updated file to disk. When run with no arguments in a TTY, the command SHALL instead launch an interactive selector (see `remove-interactive` capability).

#### Scenario: Dependency is removed from melon.yml
- **WHEN** `melon.yaml` contains `alice/pdf-skill: "^1.3.0"` and the user runs `mln remove alice/pdf-skill`
- **THEN** `melon.yaml` SHALL no longer contain an entry for `alice/pdf-skill` after the command completes

#### Scenario: No argument in TTY launches interactive mode
- **WHEN** `mln remove` is run with no arguments in a TTY
- **THEN** the command SHALL launch the interactive multi-select selector instead of exiting with an argument error
