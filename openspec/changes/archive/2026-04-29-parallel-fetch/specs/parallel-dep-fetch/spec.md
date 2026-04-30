## ADDED Requirements

### Requirement: Dependencies are fetched concurrently
During `melon install` (and commands that invoke it: `add`, `update`, `search`), the system SHALL fetch all resolved dependencies concurrently rather than sequentially, using goroutines bounded by a concurrency semaphore.

#### Scenario: Multiple dependencies are fetched in parallel
- **WHEN** `melon install` resolves two or more dependencies
- **THEN** the fetcher SHALL initiate multiple `git clone` operations concurrently, limited to at most 4 simultaneous fetches

#### Scenario: Single dependency is unaffected
- **WHEN** `melon install` resolves exactly one dependency
- **THEN** the fetch MUST complete successfully, identically to the sequential implementation

### Requirement: Lock file order is preserved after parallel fetch
The system SHALL produce a `melon.lock` with dependencies listed in the same deterministic order regardless of the order in which goroutines complete their fetches.

#### Scenario: Lock file entries match resolved order
- **WHEN** dependencies `A`, `B`, `C` are resolved in that order and fetched concurrently
- **THEN** `melon.lock` MUST list them as `A`, `B`, `C` regardless of which fetch finished first

### Requirement: A single fetch failure aborts the install
If any concurrent fetch fails, the system SHALL return an error and abort the install. Successful fetches that completed before the error was detected are left in the store (idempotency ensures they will be reused on a subsequent run).

#### Scenario: One of N concurrent fetches fails
- **WHEN** two deps are being fetched concurrently and one returns an error
- **THEN** `melon install` MUST exit with a non-zero status and report the error
- **THEN** the overall result MUST NOT produce a partial `melon.lock`

### Requirement: Concurrency is bounded to prevent resource exhaustion
The system SHALL limit the number of simultaneous `git clone` operations to a fixed maximum (4), to avoid GitHub rate limiting and file-descriptor exhaustion.

#### Scenario: More than 4 dependencies are installed
- **WHEN** `melon install` resolves 5 or more dependencies
- **THEN** at most 4 `git clone` processes MUST be running simultaneously at any point

### Requirement: TTY progress reporting remains correct under concurrent fetches
In TTY mode, each completed fetch (success or failure) SHALL send a progress event to the bubbletea program. Events from concurrent goroutines MUST NOT corrupt the TUI state.

#### Scenario: Progress bar advances for each completed fetch
- **WHEN** dependencies are fetched concurrently in TTY mode
- **THEN** the progress bar MUST advance once per completed fetch, reaching 100% when all fetches are done
