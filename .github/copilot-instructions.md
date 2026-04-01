# Github Copilot Memory Bank

Memory resets between sessions. Rely on Memory Bank files in `.github/instructions/` to understand the project.

## File Structure

**Global** (`applyTo: "**"` — loaded every prompt):
- `project-overview.instructions.md` — what the project is, 5 services, ports, scope
- `active-context.instructions.md` — current focus, decisions, patterns, next steps

**Per-service** (loaded only when editing that service):
- `identity-service.instructions.md` — folder structure, DB schema, dependencies, patterns
- `user-profile-service.instructions.md` — same
- `saving-service.instructions.md` — same
- `payment-service.instructions.md` — same
- `bff-service.instructions.md` — same

**Skills** (loaded on demand):
- `go-unit-test` — unit test patterns, harness, coverage
- `bankease-architecture` — cross-service architecture, Docker Compose, routing, design decisions
- `project-progress` — completed items, remaining work, known issues, decisions log

## Workflows

- **Start of task**: instructions are auto-loaded based on `applyTo` scope. No manual read needed.
- **update memory bank**: Update `active-context.instructions.md`. For progress updates, edit `.github/skills/project-progress/SKILL.md`.
- **After significant changes**: Update relevant instruction files (per-service or global).

**Legacy files** (still present, will be removed): `product-context`, `project-brief`, `system-patterns`, `tech-context`, `progress` — content migrated to new files above.