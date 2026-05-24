// Command lis is a fake MCP server simulating a Laboratory Information System.
//
// It exposes tools for querying lab results, pending orders, and critical values
// via the MCP JSON-RPC 2.0 protocol over stdio.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
)

// --- JSON-RPC types (duplicated intentionally — fake servers are standalone) ---

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

// --- LIS Data ---

type labResult struct {
	TestName  string  `json:"test_name"`
	Value     float64 `json:"value"`
	Unit      string  `json:"unit"`
	RefLow    float64 `json:"ref_low"`
	RefHigh   float64 `json:"ref_high"`
	Abnormal  bool    `json:"abnormal"`
	Critical  bool    `json:"critical"`
	Timestamp string  `json:"timestamp"`
}

type labPanel struct {
	PatientID string      `json:"patient_id"`
	Collected string      `json:"collected_at"`
	Results   []labResult `json:"results"`
}

var labPanels = map[string]labPanel{
	"P001": {
		PatientID: "P001", Collected: "2026-05-13T06:00:00Z",
		Results: []labResult{
			{"WBC", 18.5, "10^9/L", 4.0, 11.0, true, true, "2026-05-13T06:45:00Z"},
			{"Hemoglobin", 98, "g/L", 120, 160, true, false, "2026-05-13T06:45:00Z"},
			{"Platelets", 145, "10^9/L", 150, 400, true, false, "2026-05-13T06:45:00Z"},
			{"CRP", 186, "mg/L", 0, 10, true, true, "2026-05-13T06:45:00Z"},
			{"Procalcitonin", 3.8, "ng/mL", 0, 0.5, true, true, "2026-05-13T06:45:00Z"},
			{"Creatinine", 142, "umol/L", 60, 110, true, false, "2026-05-13T06:45:00Z"},
			{"Lactate", 4.2, "mmol/L", 0.5, 2.0, true, true, "2026-05-13T06:45:00Z"},
			{"Blood Culture", 1, "positive/negative", 0, 0, true, true, "2026-05-13T06:45:00Z"},
		},
	},
	"P002": {
		PatientID: "P002", Collected: "2026-05-13T06:00:00Z",
		Results: []labResult{
			{"WBC", 8.2, "10^9/L", 4.0, 11.0, false, false, "2026-05-13T06:45:00Z"},
			{"Hemoglobin", 108, "g/L", 120, 160, true, false, "2026-05-13T06:45:00Z"},
			{"Troponin I", 0.8, "ng/mL", 0, 0.04, true, true, "2026-05-13T06:45:00Z"},
			{"BNP", 620, "pg/mL", 0, 100, true, true, "2026-05-13T06:45:00Z"},
			{"CK-MB", 15, "U/L", 0, 25, false, false, "2026-05-13T06:45:00Z"},
			{"INR", 1.4, "ratio", 0.8, 1.2, true, false, "2026-05-13T06:45:00Z"},
		},
	},
	"P003": {
		PatientID: "P003", Collected: "2026-05-13T06:00:00Z",
		Results: []labResult{
			{"Troponin I", 12.5, "ng/mL", 0, 0.04, true, true, "2026-05-13T06:45:00Z"},
			{"CK-MB", 85, "U/L", 0, 25, true, true, "2026-05-13T06:45:00Z"},
			{"LDL", 4.8, "mmol/L", 0, 3.4, true, false, "2026-05-13T06:45:00Z"},
			{"BNP", 890, "pg/mL", 0, 100, true, true, "2026-05-13T06:45:00Z"},
			{"Potassium", 3.2, "mmol/L", 3.5, 5.0, true, false, "2026-05-13T06:45:00Z"},
		},
	},
	"P004": {
		PatientID: "P004", Collected: "2026-05-13T06:00:00Z",
		Results: []labResult{
			{"WBC", 14.2, "10^9/L", 4.0, 11.0, true, false, "2026-05-13T06:45:00Z"},
			{"CRP", 145, "mg/L", 0, 10, true, true, "2026-05-13T06:45:00Z"},
			{"Procalcitonin", 1.2, "ng/mL", 0, 0.5, true, false, "2026-05-13T06:45:00Z"},
			{"D-Dimer", 2.8, "mg/L FEU", 0, 0.5, true, true, "2026-05-13T06:45:00Z"},
			{"PaO2", 58, "mmHg", 80, 100, true, true, "2026-05-13T06:45:00Z"},
			{"PaCO2", 52, "mmHg", 35, 45, true, false, "2026-05-13T06:45:00Z"},
		},
	},
	"P005": {
		PatientID: "P005", Collected: "2026-05-13T06:00:00Z",
		Results: []labResult{
			{"Glucose", 15.2, "mmol/L", 3.9, 6.1, true, true, "2026-05-13T06:45:00Z"},
			{"HbA1c", 9.2, "%", 4, 6, true, false, "2026-05-13T06:45:00Z"},
			{"Creatinine", 185, "umol/L", 60, 110, true, true, "2026-05-13T06:45:00Z"},
			{"eGFR", 32, "mL/min", 60, 120, true, true, "2026-05-13T06:45:00Z"},
			{"PT", 16.5, "seconds", 11, 13.5, true, false, "2026-05-13T06:45:00Z"},
		},
	},
}

// --- Handlers ---

func handleInitialize(id any) jsonRPCResponse {
	return jsonRPCResponse{
		JSONRPC: "2.0", ID: id,
		Result: map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": "fake-lis", "version": "1.0.0"},
		},
	}
}

func handleToolsList(id any) jsonRPCResponse {
	tools := []map[string]any{
		{
			"name":        "query_labs",
			"description": "Query laboratory results for a patient. Returns all lab panels with reference ranges and abnormal flags.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"patient_id": map[string]any{"type": "string", "description": "Patient ID"},
				},
				"required": []string{"patient_id"},
			},
		},
		{
			"name":        "get_lab_results",
			"description": "Get latest laboratory results for a patient with reference ranges and abnormal/critical flags",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"patient_id": map[string]any{"type": "string", "description": "Patient ID"},
				},
				"required": []string{"patient_id"},
			},
		},
		{
			"name":        "get_critical_values",
			"description": "Get critical/panic lab values that require immediate clinical attention",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"patient_id": map[string]any{"type": "string", "description": "Patient ID"},
				},
				"required": []string{"patient_id"},
			},
		},
		{
			"name":        "get_trend",
			"description": "Get trend data for a specific lab test over time",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"patient_id": map[string]any{"type": "string"},
					"test_name":  map[string]any{"type": "string", "description": "Lab test name (e.g., CRP, WBC)"},
					"hours":      map[string]any{"type": "integer", "description": "Hours to look back (default: 72)"},
				},
				"required": []string{"patient_id", "test_name"},
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
	case "query_labs":
		return queryLabs(id, callParams.Arguments)
	case "get_lab_results":
		return queryLabs(id, callParams.Arguments)
	case "get_critical_values":
		return getCriticalValues(id, callParams.Arguments)
	case "get_trend":
		return getTrend(id, callParams.Arguments)
	default:
		return errResp(id, -32601, "unknown tool: "+callParams.Name)
	}
}

func queryLabs(id any, args map[string]any) jsonRPCResponse {
	pid, _ := args["patient_id"].(string)
	if pid == "" {
		return errResp(id, -32602, "patient_id required")
	}
	panel, ok := labPanels[pid]
	if !ok {
		return errResp(id, -32001, "no labs found for patient: "+pid)
	}
	return textResp(id, panel)
}

func getCriticalValues(id any, args map[string]any) jsonRPCResponse {
	pid, _ := args["patient_id"].(string)
	if pid == "" {
		return errResp(id, -32602, "patient_id required")
	}
	panel, ok := labPanels[pid]
	if !ok {
		return errResp(id, -32001, "no labs found for patient: "+pid)
	}
	var critical []labResult
	for _, r := range panel.Results {
		if r.Critical {
			critical = append(critical, r)
		}
	}
	if len(critical) == 0 {
		return textResp(id, map[string]string{"patient_id": pid, "status": "no critical values"})
	}
	return textResp(id, map[string]any{"patient_id": pid, "critical_values": critical})
}

type trendPoint struct {
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
}

func getTrend(id any, args map[string]any) jsonRPCResponse {
	pid, _ := args["patient_id"].(string)
	if pid == "" {
		return errResp(id, -32602, "patient_id required")
	}
	testName, _ := args["test_name"].(string)
	if testName == "" {
		return errResp(id, -32602, "test_name required")
	}
	hours := 72
	if h, ok := args["hours"].(float64); ok {
		hours = int(h)
	}

	panel, ok := labPanels[pid]
	if !ok {
		return errResp(id, -32001, "no labs found for patient: "+pid)
	}

	// Generate fake trend data from the current value
	var current *labResult
	for _, r := range panel.Results {
		if r.TestName == testName {
			current = &r
			break
		}
	}
	if current == nil {
		return errResp(id, -32001, fmt.Sprintf("test %s not found for patient %s", testName, pid))
	}

	now := time.Now()
	points := make([]trendPoint, 0, hours/8+1)
	for i := hours; i >= 0; i -= 8 {
		// Simple variation: slight random-ish offset based on index
		variation := current.Value * (0.8 + 0.4*float64(i%5)/4)
		points = append(points, trendPoint{
			Timestamp: now.Add(-time.Duration(i) * time.Hour).Format(time.RFC3339),
			Value:     variation,
		})
	}
	sort.Slice(points, func(i, j int) bool { return points[i].Timestamp < points[j].Timestamp })

	return textResp(id, map[string]any{
		"patient_id": pid,
		"test_name":  testName,
		"ref_range":  map[string]float64{"low": current.RefLow, "high": current.RefHigh},
		"trend":      points,
	})
}

// --- Helpers ---

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
