// Command vitals is a fake MCP server simulating a Vitals Monitoring System.
//
// It exposes tools for querying current vital signs, vital sign trends,
// and alerts for patients via the MCP JSON-RPC 2.0 protocol over stdio.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"
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

// --- Vitals Data ---

type vitalSigns struct {
	PatientID       string  `json:"patient_id"`
	HeartRate       int     `json:"heart_rate"`
	SystolicBP      int     `json:"systolic_bp"`
	DiastolicBP     int     `json:"diastolic_bp"`
	MeanAP          int     `json:"mean_arterial_pressure"`
	RespRate        int     `json:"respiratory_rate"`
	SpO2            int     `json:"spo2"`
	Temperature     float64 `json:"temperature"`
	Consciousness    string  `json:"consciousness"`
	OnVentilator    bool    `json:"on_ventilator"`
	VentilatorMode  string  `json:"ventilator_mode,omitempty"`
	FiO2            int     `json:"fio2,omitempty"`
	Peep            int     `json:"peep,omitempty"`
	TidalVolume     int     `json:"tidal_volume_ml,omitempty"`
	UrineOutput     int     `json:"urine_output_ml_hr"`
	PainScore       int     `json:"pain_score"`
	Timestamp       string  `json:"timestamp"`
}

type vitalAlert struct {
	PatientID string `json:"patient_id"`
	Type      string `json:"alert_type"`
	Message   string `json:"message"`
	Severity  string `json:"severity"`
	Timestamp string `json:"timestamp"`
}

var currentVitals = map[string]vitalSigns{
	"P001": {
		PatientID: "P001", HeartRate: 112, SystolicBP: 88, DiastolicBP: 52,
		MeanAP: 64, RespRate: 28, SpO2: 91, Temperature: 38.9,
		Consciousness: "Alert", OnVentilator: false,
		UrineOutput: 20, PainScore: 4, Timestamp: "2026-05-13T08:00:00Z",
	},
	"P002": {
		PatientID: "P002", HeartRate: 78, SystolicBP: 118, DiastolicBP: 72,
		MeanAP: 87, RespRate: 18, SpO2: 97, Temperature: 36.8,
		Consciousness: "Alert", OnVentilator: false,
		UrineOutput: 55, PainScore: 3, Timestamp: "2026-05-13T08:00:00Z",
	},
	"P003": {
		PatientID: "P003", HeartRate: 95, SystolicBP: 142, DiastolicBP: 88,
		MeanAP: 106, RespRate: 22, SpO2: 94, Temperature: 37.2,
		Consciousness: "Drowsy", OnVentilator: false,
		UrineOutput: 35, PainScore: 2, Timestamp: "2026-05-13T08:00:00Z",
	},
	"P004": {
		PatientID: "P004", HeartRate: 125, SystolicBP: 95, DiastolicBP: 58,
		MeanAP: 70, RespRate: 32, SpO2: 88, Temperature: 38.4,
		Consciousness: "Alert", OnVentilator: true,
		VentilatorMode: "SIMV", FiO2: 60, Peep: 12, TidalVolume: 400,
		UrineOutput: 15, PainScore: 5, Timestamp: "2026-05-13T08:00:00Z",
	},
	"P005": {
		PatientID: "P005", HeartRate: 68, SystolicBP: 165, DiastolicBP: 95,
		MeanAP: 118, RespRate: 16, SpO2: 96, Temperature: 36.5,
		Consciousness: "Unresponsive", OnVentilator: false,
		UrineOutput: 25, PainScore: 0, Timestamp: "2026-05-13T08:00:00Z",
	},
}

var activeAlerts = map[string][]vitalAlert{
	"P001": {
		{"P001", "hypotension", "SBP < 90 mmHg", "critical", "2026-05-13T08:00:00Z"},
		{"P001", "tachycardia", "HR > 110 bpm", "warning", "2026-05-13T08:00:00Z"},
		{"P001", "fever", "Temp > 38.5°C", "warning", "2026-05-13T08:00:00Z"},
		{"P001", "low_spo2", "SpO2 < 92%", "warning", "2026-05-13T08:00:00Z"},
	},
	"P003": {
		{"P003", "hypertension", "SBP > 140 mmHg", "warning", "2026-05-13T08:00:00Z"},
	},
	"P004": {
		{"P004", "tachycardia", "HR > 110 bpm", "warning", "2026-05-13T08:00:00Z"},
		{"P004", "low_spo2", "SpO2 < 90% on vent", "critical", "2026-05-13T08:00:00Z"},
		{"P004", "oliguria", "Urine < 30 mL/hr", "warning", "2026-05-13T08:00:00Z"},
	},
	"P005": {
		{"P005", "hypertension", "SBP > 160 mmHg", "critical", "2026-05-13T08:00:00Z"},
		{"P005", "unconscious", "GCS likely decreased", "critical", "2026-05-13T08:00:00Z"},
	},
}

// --- Handlers ---

func handleInitialize(id any) jsonRPCResponse {
	return jsonRPCResponse{
		JSONRPC: "2.0", ID: id,
		Result: map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{"tools": map[string]any{}},
			"serverInfo":      map[string]any{"name": "fake-vitals", "version": "1.0.0"},
		},
	}
}

func handleToolsList(id any) jsonRPCResponse {
	tools := []map[string]any{
		{
			"name":        "query_vitals",
			"description": "Query current vital signs for a patient including ventilator parameters if applicable",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"patient_id": map[string]any{"type": "string"},
				},
				"required": []string{"patient_id"},
			},
		},
		{
			"name":        "get_vitals",
			"description": "Get current vital signs for a patient including heart rate, blood pressure, respiratory rate, SpO2, temperature",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"patient_id": map[string]any{"type": "string", "description": "Patient ID"},
				},
				"required": []string{"patient_id"},
			},
		},
		{
			"name":        "get_alerts",
			"description": "Get active vital sign alerts for a patient or all patients",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"patient_id": map[string]any{"type": "string", "description": "Filter by patient. Empty = all alerts."},
				},
			},
		},
		{
			"name":        "get_vital_trend",
			"description": "Get trend data for a specific vital sign over the last N hours",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"patient_id":  map[string]any{"type": "string"},
					"vital_sign":  map[string]any{"type": "string", "description": "One of: heart_rate, systolic_bp, spo2, respiratory_rate, temperature"},
					"hours":       map[string]any{"type": "integer", "description": "Hours to look back (default: 24)"},
				},
				"required": []string{"patient_id", "vital_sign"},
			},
		},
	}
	return jsonRPCResponse{JSONRPC: "2.0", ID: id, Result: map[string]any{"tools": tools}}
}

func handleToolsCall(id any, params json.RawMessage) jsonRPCResponse {
	var cp struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments,omitempty"`
	}
	if err := json.Unmarshal(params, &cp); err != nil {
		return errResp(id, -32602, "invalid params")
	}
	switch cp.Name {
	case "query_vitals":
		return queryVitals(id, cp.Arguments)
	case "get_vitals":
		return queryVitals(id, cp.Arguments)
	case "get_alerts":
		return getAlerts(id, cp.Arguments)
	case "get_vital_trend":
		return getVitalTrend(id, cp.Arguments)
	default:
		return errResp(id, -32601, "unknown tool: "+cp.Name)
	}
}

func queryVitals(id any, args map[string]any) jsonRPCResponse {
	pid, _ := args["patient_id"].(string)
	if pid == "" {
		return errResp(id, -32602, "patient_id required")
	}
	v, ok := currentVitals[pid]
	if !ok {
		return errResp(id, -32001, "patient not found: "+pid)
	}
	return textResp(id, v)
}

func getAlerts(id any, args map[string]any) jsonRPCResponse {
	pid, _ := args["patient_id"].(string)
	if pid != "" {
		alerts, ok := activeAlerts[pid]
		if !ok {
			return textResp(id, map[string]string{"patient_id": pid, "alerts": "none"})
		}
		return textResp(id, map[string]any{"patient_id": pid, "alerts": alerts})
	}
	return textResp(id, map[string]any{"all_alerts": activeAlerts})
}

type trendPoint struct {
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
}

func getVitalTrend(id any, args map[string]any) jsonRPCResponse {
	pid, _ := args["patient_id"].(string)
	if pid == "" {
		return errResp(id, -32602, "patient_id required")
	}
	vitalSign, _ := args["vital_sign"].(string)
	if vitalSign == "" {
		return errResp(id, -32602, "vital_sign required")
	}
	hours := 24
	if h, ok := args["hours"].(float64); ok {
		hours = int(h)
	}

	v, ok := currentVitals[pid]
	if !ok {
		return errResp(id, -32001, "patient not found: "+pid)
	}

	var currentVal float64
	switch vitalSign {
	case "heart_rate":
		currentVal = float64(v.HeartRate)
	case "systolic_bp":
		currentVal = float64(v.SystolicBP)
	case "spo2":
		currentVal = float64(v.SpO2)
	case "respiratory_rate":
		currentVal = float64(v.RespRate)
	case "temperature":
		currentVal = v.Temperature
	default:
		return errResp(id, -32602, "unknown vital sign: "+vitalSign)
	}

	now := time.Now()
	points := make([]trendPoint, 0, hours/2+1)
	for i := hours; i >= 0; i -= 2 {
		variation := currentVal * (0.9 + 0.2*float64(i%7)/6)
		points = append(points, trendPoint{
			Timestamp: now.Add(-time.Duration(i) * time.Hour).Format(time.RFC3339),
			Value:     variation,
		})
	}
	sort.Slice(points, func(i, j int) bool { return points[i].Timestamp < points[j].Timestamp })

	return textResp(id, map[string]any{
		"patient_id":  pid,
		"vital_sign":  vitalSign,
		"current":     currentVal,
		"trend":       points,
	})
}

// --- Helpers ---

func textResp(id any, data any) jsonRPCResponse {
	j, _ := json.Marshal(data)
	return jsonRPCResponse{
		JSONRPC: "2.0", ID: id,
		Result: map[string]any{"content": []map[string]any{{"type": "text", "text": string(j)}}},
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
