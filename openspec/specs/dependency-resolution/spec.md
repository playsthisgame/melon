### Requirement: Transitive dependencies are resolved before install
The resolver SHALL fetch each dependency's `mln.yaml` transitively via the GitHub raw content API and build a complete dependency graph before any files are fetched or placed.

#### Scenario: Direct deps only (no transitive)
- **WHEN** `mln.yaml` lists `alice/pdf-skill: "^1.0.0"` and `alice/pdf-skill`'s own `mln.yaml` has no dependencies
- **THEN** the resolved set SHALL contain exactly `alice/pdf-skill` at a version satisfying `^1.0.0`

#### Scenario: Transitive dependency is included
- **WHEN** `mln.yaml` lists `alice/pdf-skill: "^1.0.0"` and `alice/pdf-skill` itself declares `bob/base-utils: "^2.0.0"`
- **THEN** the resolved set SHALL contain both `alice/pdf-skill` and `bob/base-utils` at compatible versions

#### Scenario: Diamond dependency resolves to single version
- **WHEN** two direct deps both depend on `shared/lib` with compatible constraints (e.g. `^1.0.0` and `^1.2.0`)
- **THEN** the resolved set SHALL contain `shared/lib` exactly once, at the highest version satisfying both constraints

### Requirement: Incompatible transitive constraints produce a named error
When two packages in the resolved graph require incompatible versions of a shared dependency, the resolver SHALL return an `ErrVersionConflict` error naming the conflicting constraint sources and the versions they require.

#### Scenario: Version conflict is detected and reported
- **WHEN** package A requires `shared/lib: "^1.0.0"` and package B requires `shared/lib: "^2.0.0"`
- **THEN** resolution SHALL fail with an error message that names package A, package B, `shared/lib`, and the two incompatible constraints

### Requirement: Missing or absent transitive manifest is not an error
If a dependency does not publish its own `mln.yaml`, the resolver SHALL treat it as having no transitive dependencies and continue resolution normally.

#### Scenario: Dep with no mln.yaml resolves with no transitive deps
- **WHEN** a dependency's repository has no `mln.yaml` at the expected path
- **THEN** the resolver SHALL include only that dependency itself in the resolved set, without erroring

### Requirement: Resolution is idempotent for the same manifest inputs
Given the same `mln.yaml` and the same remote tags, the resolver SHALL always return the same `[]ResolvedDep` slice in the same order.

#### Scenario: Repeated resolution returns same result
- **WHEN** `Resolve` is called twice with an identical manifest and the remote tags have not changed
- **THEN** both calls SHALL return the same resolved set with the same pinned versions
