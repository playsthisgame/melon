## ADDED Requirements

### Requirement: melon add validates source against allowed_sources before writing to melon.yaml
When a `policy` block with `allowed_sources` is present, `melon add` SHALL check the dependency path against the allowlist before modifying `melon.yaml` or running install. If the source is not permitted, the command SHALL exit non-zero with a clear error message and leave `melon.yaml` unchanged.

#### Scenario: add blocked by policy — melon.yaml unchanged
- **WHEN** `allowed_sources` is `[github.com/my-company/*]` and the user runs `melon add github.com/public/cool-skill`
- **THEN** the command SHALL exit non-zero, print a message identifying the blocked source and the active policy, and `melon.yaml` SHALL remain unmodified

#### Scenario: add permitted by policy — proceeds normally
- **WHEN** `allowed_sources` is `[github.com/my-company/*]` and the user runs `melon add github.com/my-company/approved-skill`
- **THEN** the command SHALL proceed normally, writing the dep to `melon.yaml` and running install

#### Scenario: add with no policy — no restriction
- **WHEN** no `policy` block is present in `melon.yaml`
- **THEN** `melon add` SHALL not perform any source validation and SHALL accept any dependency path
