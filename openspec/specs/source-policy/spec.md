## ADDED Requirements

### Requirement: melon.yaml supports a policy block with an allowed_sources allowlist
The manifest SHALL support an optional top-level `policy` block containing an `allowed_sources` field — a list of glob patterns defining which dependency source paths are permitted. When the `policy` block is absent or `allowed_sources` is empty, all sources are permitted and behaviour is identical to today (backwards-compatible).

#### Scenario: No policy block — all sources permitted
- **WHEN** `melon.yaml` contains no `policy` block
- **THEN** any dependency source path SHALL be accepted by both `melon add` and `melon install`

#### Scenario: Policy block with allowed_sources present
- **WHEN** `melon.yaml` contains a `policy` block with `allowed_sources: [github.com/my-company/*]`
- **THEN** only dependency paths matching at least one pattern SHALL be accepted

### Requirement: allowed_sources patterns use glob prefix matching
Each entry in `allowed_sources` SHALL be treated as a glob pattern matched against the full dependency path (e.g. `github.com/owner/repo/path/to/skill`). A trailing `*` SHALL match any suffix. Exact entries without wildcards SHALL match only that exact path prefix.

#### Scenario: Wildcard matches all repos under an org
- **WHEN** `allowed_sources` contains `github.com/my-company/*`
- **THEN** `github.com/my-company/any-repo` and `github.com/my-company/any-repo/sub/path` SHALL both be permitted

#### Scenario: Exact entry matches only that path
- **WHEN** `allowed_sources` contains `github.com/my-company/specific-repo`
- **THEN** `github.com/my-company/specific-repo` SHALL be permitted
- **THEN** `github.com/my-company/other-repo` SHALL NOT be permitted

#### Scenario: Dependency not matching any pattern is rejected
- **WHEN** `allowed_sources` is `[github.com/my-company/*]` and the user tries to install `github.com/some-stranger/cool-skill`
- **THEN** the command SHALL fail with a clear error naming the blocked dependency and the active policy

### Requirement: melon install enforces allowed_sources before fetching
`melon install` SHALL validate every dependency declared in `melon.yaml` against `allowed_sources` before beginning any fetch. If any dependency is not permitted, the install SHALL abort with a non-zero exit code listing all blocked dependencies. No fetching or lock file writing SHALL occur.

#### Scenario: Install blocked when a dep violates policy
- **WHEN** `allowed_sources` is `[github.com/my-company/*]` and `melon.yaml` contains `github.com/public/skill: "^1.0.0"`
- **THEN** `melon install` SHALL exit non-zero and print which dependency is blocked before attempting any network requests

#### Scenario: Install proceeds when all deps satisfy policy
- **WHEN** all dependencies in `melon.yaml` match at least one `allowed_sources` pattern
- **THEN** `melon install` SHALL proceed normally

### Requirement: Policy enforcement applies to manually-edited manifests
The `allowed_sources` check in `melon install` SHALL catch dependencies added by directly editing `melon.yaml`, not only those added through `melon add`, ensuring the policy cannot be bypassed by hand-editing the manifest.

#### Scenario: Manually added dep caught at install time
- **WHEN** a developer manually adds `github.com/public/skill: "main"` to `melon.yaml` and runs `melon install`
- **THEN** the install SHALL be blocked by the policy check even though `melon add` was never used
