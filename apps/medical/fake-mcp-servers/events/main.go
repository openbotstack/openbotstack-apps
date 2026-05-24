// Command events is a fake MCP server simulating a Clinical Events system.
//
// It exposes tools for querying recent clinical events, medication changes,
// and nursing interventions via the MCP JSON-RPC 2.0 protocol over stdio.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type jsonRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id"`
	Result  any       `json:"result,omitempty"`
	Error   *rpcError `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type clinicalEvent struct {
	Timestamp string `json:"timestamp"`
	Category  string `json:"category"`
	Event     string `json:"event"`
	Detail    string `json:"detail,omitempty"`
	Severity  string `json:"severity"`
}

var eventsDB = map[string][]clinicalEvent{
	"P001": {
		{"2026-05-21T06:00:00Z", "medication", "Vancomycin 1g IV started", "Renal dose adjusted", "routine"},
		{"2026-05-21T06:30:00Z", "vitals", "Hypotensive episode", "SBP dropped to 82 mmHg, fluid bolus given", "warning"},
		{"2026-05-21T07:00:00Z", "lab", "Blood culture positive", "Gram-negative rods, sensitivity pending", "critical"},
		{"2026-05-21T07:30:00Z", "nursing", "Dressing change", "Central line site clean, no signs of infection", "routine"},
		{"2026-05-21T08:00:00Z", "medication", "Norepinephrine titrated", "Increased to 0.15 mcg/kg/min", "warning"},
		{"2026-05-21T09:00:00Z", "procedure", "Chest X-ray", "Bilateral infiltrates, no pneumothorax", "routine"},
	},
	"P002": {
		{"2026-05-21T06:00:00Z", "medication", "Heparin drip continued", "PTT target 60-80, current 72", "routine"},
		{"2026-05-21T07:00:00Z", "vitals", "Stable hemodynamics", "HR 78, BP 118/72, recovering well", "routine"},
		{"2026-05-21T08:00:00Z", "nursing", "Ambulation attempt", "Tolerated 10m with assistance", "routine"},
		{"2026-05-21T09:30:00Z", "lab", "Troponin trending down", "0.8 → 0.5 → 0.3 ng/mL over 24h", "good"},
	},
	"P003": {
		{"2026-05-21T06:00:00Z", "medication", "Amiodarone infusion started", "For atrial fibrillation with rapid ventricular response", "warning"},
		{"2026-05-21T06:30:00Z", "vitals", "Chest pain reported", "8/10, NTG given with partial relief", "critical"},
		{"2026-05-21T07:00:00Z", "procedure", "Cardiac catheterization", "RCA 95% stenosis, stent placed", "critical"},
		{"2026-05-21T08:00:00Z", "nursing", "Post-procedure monitoring", "Groin access site clean, pedal pulses intact", "routine"},
		{"2026-05-21T09:00:00Z", "lab", "Post-cath labs drawn", "Troponin pending", "routine"},
	},
	"P004": {
		{"2026-05-21T06:00:00Z", "respiratory", "Ventilator settings adjusted", "FiO2 increased to 60%, PEEP 12", "warning"},
		{"2026-05-21T07:00:00Z", "vitals", "SpO2 fluctuating", "85%-91% on current settings", "critical"},
		{"2026-05-21T07:30:00Z", "nursing", "Proning protocol initiated", "Patient positioned prone for ARDS", "warning"},
		{"2026-05-21T08:30:00Z", "respiratory", "ABG post-proning", "PaO2 improved from 58 to 72 mmHg", "good"},
		{"2026-05-21T09:00:00Z", "medication", "Antibiotics continued", "Meropenem + Azithromycin per culture results", "routine"},
	},
	"P005": {
		{"2026-05-21T06:00:00Z", "neuro", "GCS assessment", "E2V3M5 = GCS 10, decreased from 13", "critical"},
		{"2026-05-21T06:30:00Z", "medication", "Blood pressure management", "Labetalol 20mg IV for SBP > 160", "warning"},
		{"2026-05-21T07:00:00Z", "procedure", "CT head follow-up", "No hemorrhagic conversion, infarct stable", "routine"},
		{"2026-05-21T08:00:00Z", "nursing", "Stroke nursing care", "Frequent neuro checks q1h, NPO status", "warning"},
		{"2026-05-21T09:00:00Z", "lab", "Glucose management", "Insulin sliding scale, BG trending down 15.2→11.8", "routine"},
	},
}

func handleInitialize(id any) jsonRPCResponse {
	return jsonRPCResponse{
		JSONRPC: "2.0", ID: id,
		Result: map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": "fake-events", "version": "1.0.0"},
		},
	}
}

func handleToolsList(id any) jsonRPCResponse {
	tools := []map[string]any{
		{
			"name":        "get_recent_events",
			"description": "Get recent clinical events for a patient (medication changes, procedures, nursing interventions, lab alerts)",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"patient_id": map[string]any{"type": "string", "description": "Patient ID"},
					"hours":      map[string]any{"type": "integer", "description": "Hours to look back (default: 8)"},
				},
				"required": []string{"patient_id"},
			},
		},
		{
			"name":        "get_current_status",
			"description": "Get a summary of the patient's current clinical status including care level, active treatments, and key concerns",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"patient_id": map[string]any{"type": "string", "description": "Patient ID"},
				},
				"required": []string{"patient_id"},
			},
		},
	}
	return jsonRPCResponse{JSONRPC: "2.0", ID: id, Result: map[string]any{"tools": tools}}
}

func handleToolsCall(id any, params json.RawMessage) jsonRPCResponse {
	var callParams struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments,omitempty"`
	}
	if err := json.Unmarshal(params, &callParams); err != nil {
		return errResp(id, -32602, "invalid params")
	}

	switch callParams.Name {
	case "get_recent_events":
		return getRecentEvents(id, callParams.Arguments)
	case "get_current_status":
		return getCurrentStatus(id, callParams.Arguments)
	default:
		return errResp(id, -32601, "unknown tool: "+callParams.Name)
	}
}

func getRecentEvents(id any, args map[string]any) jsonRPCResponse {
	pid, _ := args["patient_id"].(string)
	if pid == "" {
		return errResp(id, -32602, "patient_id required")
	}
	events, ok := eventsDB[pid]
	if !ok {
		return errResp(id, -32001, "no events found for patient: "+pid)
	}
	return textResp(id, map[string]any{
		"patient_id":    pid,
		"events":        events,
		"total_events":  len(events),
	})
}

func getCurrentStatus(id any, args map[string]any) jsonRPCResponse {
	pid, _ := args["patient_id"].(string)
	if pid == "" {
		return errResp(id, -32602, "patient_id required")
	}
	events, ok := eventsDB[pid]
	if !ok {
		return errResp(id, -32001, "no events found for patient: "+pid)
	}

	criticalCount := 0
	warningCount := 0
	for _, e := range events {
		switch e.Severity {
		case "critical":
			criticalCount++
		case "warning":
			warningCount++
		}
	}

	status := "stable"
	if criticalCount > 0 {
		status = "critical"
	} else if warningCount > 0 {
		status = "concerning"
	}

	return textResp(id, map[string]any{
		"patient_id":       pid,
		"clinical_status":  status,
		"active_treatments": extractTreatments(events),
		"key_concerns":     extractConcerns(events),
		"recent_events":    len(events),
	})
}

func extractTreatments(events []clinicalEvent) []string {
	var treatments []string
	for _, e := range events {
		if e.Category == "medication" {
			treatments = append(treatments, e.Event)
		}
	}
	return treatments
}

func extractConcerns(events []clinicalEvent) []string {
	var concerns []string
	for _, e := range events {
		if e.Severity == "critical" || e.Severity == "warning" {
			concerns = append(concerns, e.Event)
		}
	}
	return concerns
}

func textResp(id any, data any) jsonRPCResponse {
	jsonData, _ := json.Marshal(data)
	return jsonRPCResponse{
		JSONRPC: "2.0", ID: id,
		Result: map[string]any{
			"content": []map[string]any{{"type": "text", "text": string(jsonData)}},
		},
	}
}

func errResp(id any, code int, msg string) jsonRPCResponse {
	return jsonRPCResponse{JSONRPC: "2.0", ID: id, Error: &rpcError{Code: code, Message: msg}}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var req jsonRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			r, _ := json.Marshal(errResp(nil, -32700, "parse error"))
			fmt.Println(string(r))
			continue
		}
		var resp jsonRPCResponse
		switch req.Method {
		case "initialize":
			resp = handleInitialize(req.ID)
		case "notifications/initialized":
			continue
		case "tools/list":
			resp = handleToolsList(req.ID)
		case "tools/call":
			p, _ := json.Marshal(req.Params)
			resp = handleToolsCall(req.ID, p)
		default:
			resp = errResp(req.ID, -32601, "method not found: "+req.Method)
		}
		r, _ := json.Marshal(resp)
		fmt.Println(string(r))
	}
}
