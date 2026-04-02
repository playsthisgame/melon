### Requirement: Store directory is named .melon
The local package cache directory SHALL be named `.melon/`. All commands that create, read, or reference the store directory SHALL use this name.

#### Scenario: mln init creates .melon/
- **WHEN** `mln init` runs successfully
- **THEN** a `.melon/` directory exists in the project root

#### Scenario: mln install uses .melon/ as the store
- **WHEN** `mln install` fetches a dependency
- **THEN** the dependency is stored under `.melon/<encoded-name>@<version>/`

### Requirement: Manifest file is named melon.yml
The project manifest file SHALL be named `melon.yml`. All commands that read or write the manifest SHALL use this filename.

#### Scenario: mln init writes melon.yml
- **WHEN** `mln init` runs successfully
- **THEN** a `melon.yml` file exists in the project root

#### Scenario: mln install reads melon.yml
- **WHEN** `mln install` is run in a project directory
- **THEN** it reads dependencies from `melon.yml`

#### Scenario: mln init reports melon.yml
- **WHEN** `mln init --yes` completes
- **THEN** the success message references `melon.yml`

### Requirement: Lock file is named melon.lock
The reproducible lock file SHALL be named `melon.lock`. All commands that read or write the lock file SHALL use this filename.

#### Scenario: mln install writes melon.lock
- **WHEN** `mln install` completes successfully
- **THEN** a `melon.lock` file exists in the project root

#### Scenario: mln install reads existing melon.lock for diff
- **WHEN** `mln install` is run a second time
- **THEN** it loads `melon.lock` to compute the diff output
