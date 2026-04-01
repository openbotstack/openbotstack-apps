# openbotstack-apps

Application Plane for OpenBotStack — domain skills, tools, workflows, and example applications.

## Architecture

```
apps/           Example applications (demo CLI)
skills/         Domain skill definitions (nursing)
tools/          External system connectors (EHR, analytics)
workflows/      Multi-step workflow definitions
```

## Dependencies

- `openbotstack-core` — control plane types (Skill interface, JSONSchema, ExecutionPlan)

Uses a local `replace` directive for development:

```
replace github.com/openbotstack/openbotstack-core => ../openbotstack-core
```

## Quick Start

```bash
# Run tests
make test

# Run demo
make demo

# Lint
make lint
```

## Layers

### Tools

Low-level connectors to external systems. Currently provides mock implementations.

| Tool | Package | Description |
|------|---------|-------------|
| `ehr.query_patient` | `tools/ehr` | Query patient demographics |
| `ehr.query_vitals` | `tools/ehr` | Query vital signs |
| `ehr.query_labs` | `tools/ehr` | Query lab results |
| `analytics.risk_score` | `tools/analytics` | Deterministic clinical risk score |

### Skills

Domain-specific capabilities implementing `registry/skills.Skill`.

| Skill | Package | Description |
|-------|---------|-------------|
| `nursing/query_patients` | `skills/nursing` | List patients by unit |
| `nursing/summarize_status` | `skills/nursing` | Clinical status summary |
| `nursing/generate_sbar` | `skills/nursing` | SBAR handover generation |

### Workflows

Multi-step compositions that produce `ExecutionPlan` instances.

| Workflow | Description |
|----------|-------------|
| `shift_handover` | Full nursing shift handover report |
| `patient_summary` | Single patient clinical summary |

## Contract

See [AI_CONTRACT.md](./AI_CONTRACT.md) for architectural rules.
