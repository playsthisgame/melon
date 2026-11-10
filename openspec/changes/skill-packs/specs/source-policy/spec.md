## MODIFIED Requirements

### Requirement: melon install enforces allowed_sources before fetching
`melon install` SHALL validate every dependency declared in `melon.yaml` against `allowed_sources` before beginning resolution, and SHALL validate every dependency in the resolved graph — including pack members and other transitive dependencies — after resolution and before fetching any skill content. If any dependency is not permitted, the install SHALL abort with a non-zero exit code listing all blocked dependencies; for transitive dependencies the message SHALL include the `via` chain that pulled them in. No skill fetching or lock file writing SHALL occur. Resolution MAY download dependency `melon.yaml` files before the second check.

#### Scenario: Install blocked when a dep violates policy
- **WHEN** `allowed_sources` is `[github.com/my-company/*]` and `melon.yaml` contains `github.com/public/skill: "^1.0.0"`
- **THEN** `melon install` SHALL exit non-zero and print which dependency is blocked before attempting any network requests

#### Scenario: Install proceeds when all deps satisfy policy
- **WHEN** all dependencies in the resolved graph match at least one `allowed_sources` pattern
- **THEN** `melon install` SHALL proceed normally

#### Scenario: Install blocked by a pack member
- **WHEN** `allowed_sources` is `[github.com/my-company/*]`, `melon.yaml` depends on `github.com/my-company/packs/spring-pack`, and that pack declares `github.com/public/spring-boot`
- **THEN** `melon install` SHALL exit non-zero before fetching any skill, naming `github.com/public/spring-boot` and `via github.com/my-company/packs/spring-pack`, and `melon.lock` SHALL be unchanged
