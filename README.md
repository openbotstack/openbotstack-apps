# openbotstack-apps

Application Plane for OpenBotStack — domain skills, workflows, and example applications.

## Architecture

```
apps/
  medical/           Medical scenario validation (MCP servers + manifest skills + workflows)
  healthcare/        Healthcare-specific adapters (FHIR audit mapper)
examples/
  hello-world/       Minimal Wasm skill example
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

## Application: Medical Scenario

A medical scenario validation demo featuring MCP servers, manifest skills, and workflows.

See [apps/medical/](apps/medical/) for details.

## Application: Healthcare Adapters

Industry-specific audit event mappers implementing `audit.AuditEventMapper`:

- **FHIR AuditEvent mapper** — converts audit envelopes to HL7 FHIR R4 AuditEvent resources

See [apps/healthcare/audit/](apps/healthcare/audit/) for details.

## Example: Hello World

A minimal Wasm skill compiled with `GOOS=wasip1 GOARCH=wasm`.

See [examples/hello-world/](examples/hello-world/) for details.

## Contract

See [AI_CONTRACT.md](./AI_CONTRACT.md) for architectural rules.
