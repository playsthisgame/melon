## Why

Building something real (e.g. a Spring application) usually needs several related skills, and today each one has to be discovered and added individually. Melon already resolves transitive dependencies, so a `melon.yaml` that only lists dependencies is *almost* a curated group — but there is no way to say so: the resolver assumes every dep is a skill with a `SKILL.md`, the placer would symlink the group's own directory as if it were a skill, and nothing records why a skill was installed. Making "packs" a first-class concept lets the community publish curated sets like `spring-expert-pack` that install with one `melon add`.

## What Changes

- `melon.yaml` gains an optional `type` field: `skill` (default) or `pack`. Any other value is a validation error.
- A **pure pack** (`type: pack`, no entrypoint file in its directory) is resolved and fetched but never placed into harness directories. A **hybrid pack** (`type: pack` whose entrypoint file exists, e.g. an orchestrating "expert" `SKILL.md`) is placed like a normal skill in addition to its members.
- `melon.lock` records `type` and provenance per dependency: `direct: true` when listed in the project's `melon.yaml`, and `via: [...]` listing every resolved dependency that requires it.
- `melon list` shows pack members nested under their pack; `--json` output includes `type`, `direct`, and `via`.
- `melon remove <pack>` removes the pack and prunes members that nothing else requires; members that are also direct deps or required by another pack are kept.
- A direct dependency in the user's `melon.yaml` takes precedence over a pack's constraint for the same skill, provided the pack's constraint is still satisfied; otherwise a conflict error names the pack.
- `allowed_sources` is enforced against **every** resolved dependency (including pack members and other transitive deps), after resolution and before any fetch.
- Index entries may carry `type: pack`; `melon search` labels packs, and `melon info <pack>` lists its member skills and constraints.
- `melon add` runs the full resolve/install pipeline when the added dep is a pack or declares dependencies (today it installs only the named dep, so transitive deps are silently skipped). `melon remove` does the same for packs, and refuses to remove a skill that is only installed via a pack, naming the pack.
- `melon add <pack>` prints the full member list before installing and, in a TTY, asks for confirmation (default yes; skipped with `--yes`).
- Out of scope: excluding individual pack members, `melon init --pack` scaffolding, pack publishing tooling.

## Capabilities

### New Capabilities
- `skill-packs`: the pack manifest type, pack/hybrid placement rules, and lock-file provenance (`type`, `direct`, `via`).

### Modified Capabilities
- `dependency-resolution`: resolver records the full set of parents for every dep and direct deps take precedence over pack constraints.
- `skill-placement`: pure packs are not placed; hybrid packs are.
- `source-policy`: install enforces `allowed_sources` on the full resolved graph, not only direct deps.
- `list-command`: tree display of pack members and provenance fields in JSON output.
- `remove-command`: removing a pack prunes members that are no longer required by anything.
- `skill-search`: index entries support `type: pack` and results label packs.
- `skill-info`: `melon info` on a pack lists its members.
- `add-command`: adding a pack shows its members and confirms in TTY mode.

## Impact

- **Code**: `internal/manifest` (schema + validation), `internal/resolver` (parent tracking, type propagation), `internal/lockfile` (new fields, diff/frozen comparison), `internal/placer` (skip pure packs), `internal/index` (entry type), `internal/cli` (install policy check, list, remove, add, search, info).
- **Lock file format**: additive fields (`type`, `direct`, `via`). Old lock files load fine; the first install after upgrade rewrites them with provenance. Older melon versions ignore the new fields.
- **Manifest format**: additive `type` field; existing manifests are treated as `type: skill`.
- **Behavior change**: projects with an `allowed_sources` policy may now see installs blocked by transitive deps that were previously let through silently. This is intentional (closing a policy gap) but should be called out in release notes.
- No new external dependencies.
