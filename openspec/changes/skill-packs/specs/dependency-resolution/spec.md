## ADDED Requirements

### Requirement: Resolver records every parent of each dependency
The resolver SHALL record, for each resolved dependency, whether it is a direct dependency of the root manifest and the full set of resolved dependencies that declare it — including parents encountered after the dependency was first selected. The resolved result SHALL also carry each dependency's `type` from its manifest.

#### Scenario: Diamond records both parents
- **WHEN** direct deps A and B both declare `shared/lib` with compatible constraints
- **THEN** the resolved entry for `shared/lib` SHALL list both A and B as parents and SHALL NOT be marked direct

#### Scenario: Direct dep also declared transitively
- **WHEN** the root manifest lists `shared/lib` and pack P also declares `shared/lib`
- **THEN** the resolved entry SHALL be marked direct and SHALL list P as a parent

### Requirement: Direct dependencies take precedence over pack constraints
When a dependency is listed both in the root manifest and in a pack, the version SHALL be selected from the root manifest's constraint. The pack's constraint SHALL then be checked against that version; if it is not satisfied, resolution SHALL fail with `ErrConflict`, and the error message SHALL identify the root manifest and the pack by name and suggest loosening one of the constraints.

#### Scenario: Compatible direct override
- **WHEN** the root manifest lists `spring-boot: "^1.5.0"` and `spring-expert-pack` declares `spring-boot: "^1.0.0"`, and `1.6.0` is the latest tag
- **THEN** `spring-boot` SHALL resolve to `1.6.0`

#### Scenario: Incompatible direct override
- **WHEN** the root manifest lists `spring-boot: "^2.0.0"` and `spring-expert-pack` declares `spring-boot: "^1.0.0"`
- **THEN** resolution SHALL fail with an error naming `melon.yaml`, `spring-expert-pack`, `spring-boot`, and both constraints
