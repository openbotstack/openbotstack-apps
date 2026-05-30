# Medical App API Integration Guide

This document shows how the OpenBotStack platform API is used for medical-domain
skill execution. These examples are specific to the `medical` app and its MCP
tool integrations.

## Chat Request Examples

### First-Day Clinical Note

```json
{
  "message": "Generate a first-day note for patient P001",
  "session_id": "med-session-001"
}
```

Response:

```json
{
  "session_id": "med-session-001",
  "message": "### FIRST DAY NOTE\n**Patient**: Zhang Wei (P001), 72/M, Bed ICU-01\n**Admission Date**: 2026-05-10\n**Admitting Diagnosis**: Sepsis\n...",
  "skill_used": "medical.first-day-note",
  "execution_id": "exec-abc123"
}
```

### SBAR Shift Handover

```json
{
  "message": "Prepare SBAR handover for patient P003",
  "session_id": "med-session-002"
}
```

Response:

```json
{
  "session_id": "med-session-002",
  "message": "### SBAR SHIFT HANDOVER\n**Patient**: Wang Jun (P003), 68/M, Bed CCU-05\n#### S — Situation\n...",
  "skill_used": "medical.sbar-handover",
  "execution_id": "exec-def456"
}
```

## Reasoning Trace

Retrieved via `GET /v1/execution/{id}/reasoning`. Shows the full execution chain:
plan → skill selection → MCP tool calls → output.

```json
{
  "execution_id": "exec-abc123",
  "trace": [
    { "type": "plan", "summary": "execution with 3 step(s)" },
    {
      "type": "tool_call",
      "summary": "Call get_patient_demographics",
      "input": { "patient_id": "P001" }
    },
    {
      "type": "observation",
      "summary": "Result from get_patient_demographics",
      "output": { "id": "P001", "name": "Zhang Wei", "primary_diagnosis": "Sepsis" }
    },
    {
      "type": "tool_call",
      "summary": "Call get_lab_results",
      "input": { "patient_id": "P001" }
    },
    {
      "type": "observation",
      "summary": "Result from get_lab_results",
      "output": {
        "results": [
          { "test_name": "WBC", "value": 18.5, "abnormal": true, "critical": true }
        ]
      }
    },
    { "type": "decision", "summary": "execution completed" }
  ]
}
```

## SSE Stream Example

Full streaming flow for a clinical query using MCP tools + SBAR handover skill:

```
event: progress
data: {"type":"analyzing","content":"Analyzing request..."}

event: progress
data: {"type":"loading_context","content":"Loading context..."}

event: progress
data: {"type":"planning","content":""}

event: progress
data: {"type":"step_start","content":"mcp.his.get_patient_demographics"}

event: progress
data: {"type":"step_complete","content":"mcp.his.get_patient_demographics"}

event: progress
data: {"type":"step_start","content":"mcp.vitals.get_vitals"}

event: progress
data: {"type":"step_complete","content":"mcp.vitals.get_vitals"}

event: progress
data: {"type":"step_start","content":"mcp.events.get_recent_events"}

event: progress
data: {"type":"step_complete","content":"mcp.events.get_recent_events"}

event: progress
data: {"type":"step_start","content":"medical.sbar-handover"}

event: progress
data: {"type":"token","content":"### SBAR SHIFT HANDOVER\n\n**Patient**: Chen Mei (P004), ICU-04\n..."}

event: progress
data: {"type":"step_complete","content":"medical.sbar-handover"}

event: session
data: {"session_id":"519a4850-71e6-4243-9c6b-2f788fa6d6de"}

event: done
data: {"execution_id":"df46c06f-db15-4ac2-a942-3b34b278ea1c"}
```

## MCP Tool Requirements

| Skill | MCP Tools | Permissions |
|-------|-----------|-------------|
| `medical.first-day-note` | `get_patient_demographics`, `get_diagnosis`, `get_lab_results`, `get_vitals` | `mcp:his`, `mcp:lis`, `mcp:vitals` |
| `medical.sbar-handover` | `get_current_status`, `get_recent_events`, `get_vitals` | `mcp:vitals`, `mcp:events` |

## Test Patients

| ID | Name | Age/Sex | Diagnosis |
|----|------|---------|-----------|
| P001 | Zhang Wei | 72/M | Sepsis |
| P002 | Li Na | 45/F | Post-op cardiac bypass |
| P003 | Wang Jun | 68/M | Acute MI |
| P004 | Chen Mei | 55/F | ARDS |
| P005 | Zhao Yang | 80/M | Ischemic Stroke |
