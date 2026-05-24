# openbotstack-apps

Application Plane for OpenBotStack — domain skills, workflows, and example applications.

## Architecture

```
apps/
  hospital-demo/     Hospital ICU/CCU demo (MCP servers + manifest skills + workflows)
  healthcare/        Healthcare-specific adapters (FHIR audit mapper)
```

## Dependencies

- `openbotstack-core` — control plane types (Skill interface, JSONSchema, ExecutionPlan, AuditEventMapper)

Uses a local `replace` directive for development:

```
replace github.com/openbotstack/openbotstack-core => ../openbotstack-core
```

## Quick Start

```bash
# Run tests
make test

# Lint
make lint
```

## Application: Hospital Demo

A complete hospital ICU/CCU demonstration featuring:

- **3 MCP servers** (HIS, LIS, Vitals) — mock external systems via stdio JSON-RPC
- **5 manifest-based skills** — declarative SKILL.md + manifest.yaml
- **3 workflows** — ICU round, abnormal lab investigation, admission summary
- **E2E tests** — full integration test suite

See [apps/hospital-demo/README.md](apps/hospital-demo/README.md) for details.

## Application: Healthcare Adapters

Industry-specific audit event mappers implementing `audit.AuditEventMapper`:

- **FHIR AuditEvent mapper** — converts audit envelopes to HL7 FHIR R4 AuditEvent resources

See [apps/healthcare/audit/](apps/healthcare/audit/) for details.

## Contract

See [AI_CONTRACT.md](./AI_CONTRACT.md) for architectural rules.
