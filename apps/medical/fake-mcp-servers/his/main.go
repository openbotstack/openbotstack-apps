// Command his is a fake MCP server simulating a Hospital Information System.
//
// It exposes tools for querying patient demographics, admissions, and diagnoses
// via the MCP JSON-RPC 2.0 protocol over stdio.
//
// Build: go build -o his-server .
// Run:   ./his-server (launched by OpenBotStack MCP client via stdio transport)
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// --- JSON-RPC 2.0 types ---

type jsonRPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id,omitempty"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      any         `json:"id"`
	Result  any         `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// --- HIS Data ---

type patient struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Age            int    `json:"age"`
	Gender         string `json:"gender"`
	Unit           string `json:"unit"`
	Bed            string `json:"bed"`
	AdmissionDate  string `json:"admission_date"`
	AttendingDr    string `json:"attending_dr"`
	PrimaryDiag    string `json:"primary_diagnosis"`
	SecondaryDiags []string `json:"secondary_diagnoses,omitempty"`
	Allergies      []string `json:"allergies,omitempty"`
	Isolation      string `json:"isolation,omitempty"`
	CodeStatus     string `json:"code_status"`
}

var patients = map[string]patient{
	"P001": {
		ID: "P001", Name: "Zhang Wei", Age: 72, Gender: "M",
		Unit: "ICU", Bed: "ICU-01", AdmissionDate: "2026-05-10",
		AttendingDr: "Dr. Liu", PrimaryDiag: "Sepsis",
		SecondaryDiags: []string{"Type 2 Diabetes", "Hypertension"},
		Allergies: []string{"Penicillin"},
		CodeStatus: "Full Code",
	},
	"P002": {
		ID: "P002", Name: "Li Na", Age: 45, Gender: "F",
		Unit: "ICU", Bed: "ICU-02", AdmissionDate: "2026-05-11",
		AttendingDr: "Dr. Wang", PrimaryDiag: "Post-operative monitoring (cardiac bypass)",
		SecondaryDiags: []string{"Coronary Artery Disease"},
		CodeStatus: "Full Code",
	},
	"P003": {
		ID: "P003", Name: "Wang Jun", Age: 68, Gender: "M",
		Unit: "CCU", Bed: "CCU-05", AdmissionDate: "2026-05-12",
		AttendingDr: "Dr. Chen", PrimaryDiag: "Acute Myocardial Infarction",
		SecondaryDiags: []string{"Atrial Fibrillation", "Hyperlipidemia"},
		Allergies: []string{"Aspirin", "Iodine contrast"},
		Isolation: "Contact",
		CodeStatus: "DNR",
	},
	"P004": {
		ID: "P004", Name: "Chen Mei", Age: 55, Gender: "F",
		Unit: "ICU", Bed: "ICU-04", AdmissionDate: "2026-05-09",
		AttendingDr: "Dr. Liu", PrimaryDiag: "ARDS (Acute Respiratory Distress Syndrome)",
		SecondaryDiags: []string{"Pneumonia", "COPD"},
		CodeStatus: "Full Code",
	},
	"P005": {
		ID: "P005", Name: "Zhao Yang", Age: 80, Gender: "M",
		Unit: "ICU", Bed: "ICU-06", AdmissionDate: "2026-05-08",
		AttendingDr: "Dr. Zhang", PrimaryDiag: "Stroke (Ischemic)",
		SecondaryDiags: []string{"Atrial Fibrillation", "Type 2 Diabetes", "Chronic Kidney Disease Stage 3"},
		Allergies: []string{"Sulfa drugs"},
		CodeStatus: "DNI",
	},
}

// --- MCP Protocol Handlers ---

func handleInitialize(id any) jsonRPCResponse {
	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo": map[string]any{
				"name":    "fake-his",
				"version": "1.0.0",
			},
		},
	}
}

func handleToolsList(id any) jsonRPCResponse {
	tools := []map[string]any{
		{
			"name":        "query_patient",
			"description": "Query patient demographics, admission info, and diagnoses from the Hospital Information System",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"patient_id": map[string]any{"type": "string", "description": "Patient ID (e.g., P001)"},
				},
				"required": []string{"patient_id"},
			},
		},
		{
			"name":        "list_patients",
			"description": "List all patients in a specific unit or all ICU/CCU patients",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"unit": map[string]any{"type": "string", "description": "Unit filter (ICU, CCU). Empty = all."},
				},
			},
		},
		{
			"name":        "get_code_status",
			"description": "Get the code status (resuscitation preference) for a patient",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"patient_id": map[string]any{"type": "string"},
				},
				"required": []string{"patient_id"},
			},
		},
		{
				"name":        "get_patient_demographics",
				"description": "Get patient demographics: name, age, gender, unit, bed, admission date, attending physician",
				"inputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"patient_id": map[string]any{"type": "string", "description": "Patient ID"},
					},
					"required": []string{"patient_id"},
				},
		},
		{
				"name":        "get_diagnosis",
				"description": "Get patient diagnoses: primary diagnosis, comorbidities, allergies, code status",
				"inputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"patient_id": map[string]any{"type": "string", "description": "Patient ID"},
					},
					"required": []string{"patient_id"},
				},
		},
	}
	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  map[string]any{"tools": tools},
	}
}

func handleToolsCall(id any, params json.RawMessage) jsonRPCResponse {
	var callParams struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments,omitempty"`
	}
	if err := json.Unmarshal(params, &callParams); err != nil {
		return errorResponse(id, -32602, "invalid params: "+err.Error())
	}

	switch callParams.Name {
	case "query_patient":
		return queryPatient(id, callParams.Arguments)
	case "list_patients":
		return listPatients(id, callParams.Arguments)
	case "get_code_status":
		return getCodeStatus(id, callParams.Arguments)
	case "get_patient_demographics":
		return getPatientDemographics(id, callParams.Arguments)
	case "get_diagnosis":
		return getDiagnosis(id, callParams.Arguments)
	default:
		return errorResponse(id, -32601, "unknown tool: "+callParams.Name)
	}
}

func queryPatient(id any, args map[string]any) jsonRPCResponse {
	pid, _ := args["patient_id"].(string)
	if pid == "" {
		return errorResponse(id, -32602, "patient_id required")
	}
	p, ok := patients[pid]
	if !ok {
		return errorResponse(id, -32001, "patient not found: "+pid)
	}
	return toolTextResponse(id, p)
}

func listPatients(id any, args map[string]any) jsonRPCResponse {
	unit, _ := args["unit"].(string)
	var result []patient
	for _, p := range patients {
		if unit == "" || strings.EqualFold(p.Unit, unit) {
			result = append(result, p)
		}
	}
	return toolTextResponse(id, result)
}

func getCodeStatus(id any, args map[string]any) jsonRPCResponse {
	pid, _ := args["patient_id"].(string)
	if pid == "" {
		return errorResponse(id, -32602, "patient_id required")
	}
	p, ok := patients[pid]
	if !ok {
		return errorResponse(id, -32001, "patient not found: "+pid)
	}
	return toolTextResponse(id, map[string]string{
		"patient_id":   pid,
		"patient_name": p.Name,
		"code_status":  p.CodeStatus,
	})
}

// --- Additional tool handlers ---

func getPatientDemographics(id any, args map[string]any) jsonRPCResponse {
	pid, _ := args["patient_id"].(string)
	if pid == "" {
		return errorResponse(id, -32602, "patient_id required")
	}
	p, ok := patients[pid]
	if !ok {
		return errorResponse(id, -32001, "patient not found: "+pid)
	}
	return toolTextResponse(id, map[string]any{
		"patient_id":      pid,
		"name":            p.Name,
		"age":             p.Age,
		"gender":          p.Gender,
		"unit":            p.Unit,
		"bed":             p.Bed,
		"admission_date":  p.AdmissionDate,
		"attending_dr":    p.AttendingDr,
	})
}

func getDiagnosis(id any, args map[string]any) jsonRPCResponse {
	pid, _ := args["patient_id"].(string)
	if pid == "" {
		return errorResponse(id, -32602, "patient_id required")
	}
	p, ok := patients[pid]
	if !ok {
		return errorResponse(id, -32001, "patient not found: "+pid)
	}
	return toolTextResponse(id, map[string]any{
		"patient_id":            pid,
		"primary_diagnosis":     p.PrimaryDiag,
		"secondary_diagnoses":   p.SecondaryDiags,
		"allergies":             p.Allergies,
		"code_status":           p.CodeStatus,
		"isolation":             p.Isolation,
	})
}

// --- Helpers ---

func toolTextResponse(id any, data any) jsonRPCResponse {
	jsonData, _ := json.Marshal(data)
	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]any{
			"content": []map[string]any{
				{"type": "text", "text": string(jsonData)},
			},
		},
	}
}

func errorResponse(id any, code int, msg string) jsonRPCResponse {
	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &rpcError{Code: code, Message: msg},
	}
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
			resp, _ := json.Marshal(errorResponse(nil, -32700, "parse error"))
			fmt.Println(string(resp))
			continue
		}

		var resp jsonRPCResponse
		switch req.Method {
		case "initialize":
			resp = handleInitialize(req.ID)
		case "notifications/initialized":
			continue // notification, no response
		case "tools/list":
			resp = handleToolsList(req.ID)
		case "tools/call":
			rawParams, _ := json.Marshal(req.Params)
			resp = handleToolsCall(req.ID, rawParams)
		default:
			resp = errorResponse(req.ID, -32601, "method not found: "+req.Method)
		}

		respBytes, _ := json.Marshal(resp)
		fmt.Println(string(respBytes))
	}
}
