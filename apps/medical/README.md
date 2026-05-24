# Medical Scenario Validation

Demonstrates OpenBotStack as an **embedded AI execution platform** for healthcare, validating the full execution chain: planner → skill → MCP tool → result.

## Architecture

```
External System (HIS/LIS/Vitals/Events)
        ↕ MCP (JSON-RPC over stdio)
OpenBotStack Runtime
        ↕ Declarative Skills
Planner → Skill → MCP Tools → Structured Output
```

**Key principle**: OpenBotStack is a platform, not an application. External systems integrate via API. Skills are declarative configurations, not code.

## Skills

### `medical.first-day-note`

Generates structured first-day clinical notes by aggregating data from multiple systems.

- **Input**: `patient_id` (string)
- **MCP tools used**: `get_patient_demographics`, `get_diagnosis`, `get_lab_results`, `get_vitals`
- **Permissions required**: `mcp:his`, `mcp:lis`, `mcp:vitals`

```bash
curl -X POST http://localhost:8080/v1/chat \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-key" \
  -d '{"message": "Generate a first-day note for patient P001"}'
```

### `medical.sbar-handover`

Generates SBAR shift handover documents with current status and recent events.

- **Input**: `patient_id` (string)
- **MCP tools used**: `get_current_status`, `get_recent_events`, `get_vitals`
- **Permissions required**: `mcp:vitals`, `mcp:events`

```bash
curl -X POST http://localhost:8080/v1/chat \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-key" \
  -d '{"message": "Prepare SBAR handover for patient P003"}'
```

## MCP Servers (Fake)

Four fake MCP servers simulate external healthcare systems for development:

| Server | Binary | Tools | Description |
|--------|--------|-------|-------------|
| HIS | `his-server` | `query_patient`, `get_patient_demographics`, `get_diagnosis`, `list_patients`, `get_code_status` | Patient demographics, diagnoses, admissions |
| LIS | `lis-server` | `query_labs`, `get_lab_results`, `get_critical_values`, `get_trend` | Laboratory results with reference ranges |
| Vitals | `vitals-server` | `query_vitals`, `get_vitals`, `get_alerts`, `get_vital_trend` | Vital signs with ventilator parameters |
| Events | `events-server` | `get_recent_events`, `get_current_status` | Clinical events, medication changes, procedures |

### Build

```bash
# Build all fake MCP servers
for dir in fake-mcp-servers/*/; do
  (cd "$dir" && go build -o $(basename "$dir")-server .)
done
```

### Register via Admin API

```bash
# Register HIS server
curl -X POST http://localhost:8080/v1/admin/mcp/servers \
  -H "Content-Type: application/json" \
  -H "X-API-Key: admin-key" \
  -d '{
    "id": "his",
    "name": "HIS",
    "transport": "stdio",
    "command": "./fake-mcp-servers/his/his-server"
  }'

# Register LIS server
curl -X POST http://localhost:8080/v1/admin/mcp/servers \
  -d '{"id": "lis", "name": "LIS", "transport": "stdio", "command": "./fake-mcp-servers/lis/lis-server"}'

# Register Vitals server
curl -X POST http://localhost:8080/v1/admin/mcp/servers \
  -d '{"id": "vitals", "name": "Vitals", "transport": "stdio", "command": "./fake-mcp-servers/vitals/vitals-server"}'

# Register Events server
curl -X POST http://localhost:8080/v1/admin/mcp/servers \
  -d '{"id": "events", "name": "Events", "transport": "stdio", "command": "./fake-mcp-servers/events/events-server"}'
```

### Discover Tools

```bash
# List all MCP tools
curl http://localhost:8080/v1/admin/mcp/servers/his/tools

# List all capabilities (skills + MCP tools)
curl http://localhost:8080/v1/admin/capabilities
```

## Demo Client

A single-file HTML developer tool that demonstrates external system integration.

```bash
# Open in browser
open demo-client/index.html
# Or serve via any static server
python3 -m http.server 3000 --directory demo-client
```

### Features

- Streaming chat via `POST /v1/chat/stream` (SSE)
- Execution ID capture from stream events
- Reasoning trace retrieval via `GET /v1/execution/{id}/reasoning`
- Visualization: plan steps, skill calls, MCP tool calls, input/output, errors
- Metadata panel: execution ID, duration, step count, stop reason
- Quick actions for common medical queries

## Test Patients

| ID | Name | Age | Unit | Primary Diagnosis |
|----|------|-----|------|-------------------|
| P001 | Zhang Wei | 72/M | ICU-01 | Sepsis |
| P002 | Li Na | 45/F | ICU-02 | Post-operative (cardiac bypass) |
| P003 | Wang Jun | 68/M | CCU-05 | Acute Myocardial Infarction |
| P004 | Chen Mei | 55/F | ICU-04 | ARDS |
| P005 | Zhao Yang | 80/M | ICU-06 | Stroke (Ischemic) |

## Validation Checklist

- [ ] Planner correctly selects `medical.first-day-note` when asked for first-day note
- [ ] Planner correctly selects `medical.sbar-handover` when asked for SBAR handover
- [ ] MCP tools are invoked (get_patient_demographics, get_diagnosis, get_lab_results, get_vitals, get_current_status, get_recent_events)
- [ ] Output is structured and medically usable
- [ ] Reasoning trace shows complete chain: plan → tool calls → observations → decision
- [ ] Execution ID is returned and trace is retrievable
- [ ] Demo client displays streaming output and reasoning visualization

## Architecture Compliance

This demo follows OpenBotStack architectural boundaries:

- **Skills are in apps/** — declarative configuration, no code
- **MCP is runtime-managed** — configured via admin API, not hardcoded
- **No workflow DSL** — planner decides execution order
- **No UI in runtime/web** — demo-client is a standalone developer tool
- **Skills reference logical tool names** — no direct MCP server coupling
