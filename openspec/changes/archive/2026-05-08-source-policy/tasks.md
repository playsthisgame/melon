## 1. Schema

- [x] 1.1 Add `PolicyConfig` struct to `internal/manifest/schema.go` with `AllowedSources []string` field and `yaml:"allowed_sources,omitempty"` tag
- [x] 1.2 Add `Policy *PolicyConfig` field to the `Manifest` struct with `yaml:"policy,omitempty"` tag
- [x] 1.3 Add `TestPolicy_RoundTrip` and `TestPolicy_AbsentFieldRoundTrip` to `internal/manifest/manifest_test.go`

## 2. Source Matching

- [x] 2.1 Implement `matchesAllowedSources(depPath string, patterns []string) bool` in `internal/cli/` using `path.Match` glob semantics
- [x] 2.2 Add `checkSourcePolicy(m manifest.Manifest, depPaths []string) error` that collects all blocked deps and returns a single error listing them all
- [x] 2.3 Write unit tests for `matchesAllowedSources` covering: wildcard org match, exact path match, no-match, empty allowlist (permits all)

## 3. Enforcement in melon add

- [x] 3.1 In `internal/cli/add_cmd.go`, call `checkSourcePolicy` with the new dep path before writing to `melon.yaml`
- [x] 3.2 Ensure `melon.yaml` is not modified when the policy check fails
- [x] 3.3 Write a test verifying `melon add` is blocked and `melon.yaml` is unchanged when source violates policy
- [x] 3.4 Write a test verifying `melon add` proceeds normally when source matches policy

## 4. Enforcement in melon install

- [x] 4.1 In the install pipeline, call `checkSourcePolicy` with all dep paths from `melon.yaml` before beginning any resolve/fetch work
- [x] 4.2 Ensure install exits non-zero and lists ALL blocked deps (not just the first) when policy is violated
- [x] 4.3 Write a test verifying `melon install` is blocked before any network calls when a dep violates policy
- [x] 4.4 Write a test verifying `melon install` proceeds normally when all deps satisfy policy

## 5. README

- [x] 5.1 Add `policy` block to the manifest reference example in `README.md` with `allowed_sources` and inline comments explaining glob syntax
- [x] 5.2 Add a note to the `melon add` and `melon install` command docs explaining that source policy is enforced
