package main

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

func buildHIS(t *testing.T) string {
	t.Helper()
	bin := t.TempDir() + "/his-server"
	out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput()
	if err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	return bin
}

func sendHIS(t *testing.T, bin, input string) jsonRPCResponse {
	t.Helper()
	cmd := exec.Command(bin)
	cmd.Stdin = strings.NewReader(input + "\n")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run: %v\n%s", err, out)
	}
	var resp jsonRPCResponse
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(out))), &resp); err != nil {
		t.Fatalf("unmarshal: %v\noutput: %s", err, out)
	}
	return resp
}

func TestHIS_Initialize(t *testing.T) {
	bin := buildHIS(t)
	resp := sendHIS(t, bin, `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`)
	if resp.Error != nil {
		t.Fatalf("error: %s", resp.Error.Message)
	}
	result := resp.Result.(map[string]any)
	info := result["serverInfo"].(map[string]any)
	if info["name"] != "fake-his" {
		t.Errorf("name = %v, want fake-his", info["name"])
	}
}

func TestHIS_ToolsList(t *testing.T) {
	bin := buildHIS(t)
	resp := sendHIS(t, bin, `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)
	result := resp.Result.(map[string]any)
	tools := result["tools"].([]any)
	if len(tools) != 5 {
		t.Fatalf("expected 5 tools, got %d", len(tools))
	}
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		tm := tool.(map[string]any)
		toolNames[tm["name"].(string)] = true
	}
	for _, name := range []string{"query_patient", "list_patients", "get_code_status", "get_patient_demographics", "get_diagnosis"} {
		if !toolNames[name] {
			t.Errorf("missing tool: %s", name)
		}
	}
}

func TestHIS_QueryPatient(t *testing.T) {
	bin := buildHIS(t)
	resp := sendHIS(t, bin, `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"query_patient","arguments":{"patient_id":"P001"}}}`)
	if resp.Error != nil {
		t.Fatalf("error: %s", resp.Error.Message)
	}
	text := extractText(t, resp)

	var patient map[string]any
	json.Unmarshal([]byte(text), &patient)
	if patient["name"] != "Zhang Wei" {
		t.Errorf("name = %v, want Zhang Wei", patient["name"])
	}
	if patient["primary_diagnosis"] != "Sepsis" {
		t.Errorf("diagnosis = %v, want Sepsis", patient["primary_diagnosis"])
	}
}

func TestHIS_QueryPatient_NotFound(t *testing.T) {
	bin := buildHIS(t)
	resp := sendHIS(t, bin, `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"query_patient","arguments":{"patient_id":"P999"}}}`)
	if resp.Error == nil {
		t.Fatal("expected error for unknown patient")
	}
}

func TestHIS_ListPatients_ByUnit(t *testing.T) {
	bin := buildHIS(t)
	resp := sendHIS(t, bin, `{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"list_patients","arguments":{"unit":"ICU"}}}`)
	if resp.Error != nil {
		t.Fatalf("error: %s", resp.Error.Message)
	}
	text := extractText(t, resp)

	var patients []map[string]any
	json.Unmarshal([]byte(text), &patients)
	for _, p := range patients {
		if p["unit"] != "ICU" {
			t.Errorf("got non-ICU patient: %v", p["name"])
		}
	}
}

func TestHIS_GetCodeStatus(t *testing.T) {
	bin := buildHIS(t)
	resp := sendHIS(t, bin, `{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"get_code_status","arguments":{"patient_id":"P003"}}}`)
	if resp.Error != nil {
		t.Fatalf("error: %s", resp.Error.Message)
	}
	text := extractText(t, resp)

	var status map[string]any
	json.Unmarshal([]byte(text), &status)
	if status["code_status"] != "DNR" {
		t.Errorf("code_status = %v, want DNR", status["code_status"])
	}
}

func extractText(t *testing.T, resp jsonRPCResponse) string {
	t.Helper()
	result := resp.Result.(map[string]any)
	content := result["content"].([]any)
	return content[0].(map[string]any)["text"].(string)
}

func TestHIS_GetPatientDemographics(t *testing.T) {
	bin := buildHIS(t)
	resp := sendHIS(t, bin, `{"jsonrpc":"2.0","id":10,"method":"tools/call","params":{"name":"get_patient_demographics","arguments":{"patient_id":"P003"}}}`)
	if resp.Error != nil {
		t.Fatalf("get_patient_demographics error: %s", resp.Error.Message)
	}
	content := resp.Result.(map[string]any)["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)
	if !strings.Contains(text, "Wang Jun") {
		t.Errorf("expected Wang Jun in response, got: %s", text)
	}
	if !strings.Contains(text, "CCU-05") {
		t.Errorf("expected CCU-05 in response, got: %s", text)
	}
}

func TestHIS_GetPatientDemographics_EmptyID(t *testing.T) {
	bin := buildHIS(t)
	resp := sendHIS(t, bin, `{"jsonrpc":"2.0","id":11,"method":"tools/call","params":{"name":"get_patient_demographics","arguments":{"patient_id":""}}}`)
	if resp.Error == nil {
		t.Fatal("expected error for empty patient_id")
	}
	if resp.Error.Code != -32602 {
		t.Errorf("error code = %d, want -32602", resp.Error.Code)
	}
}

func TestHIS_GetDiagnosis(t *testing.T) {
	bin := buildHIS(t)
	resp := sendHIS(t, bin, `{"jsonrpc":"2.0","id":12,"method":"tools/call","params":{"name":"get_diagnosis","arguments":{"patient_id":"P001"}}}`)
	if resp.Error != nil {
		t.Fatalf("get_diagnosis error: %s", resp.Error.Message)
	}
	content := resp.Result.(map[string]any)["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)
	if !strings.Contains(text, "Sepsis") {
		t.Errorf("expected Sepsis in response, got: %s", text)
	}
	if !strings.Contains(text, "Penicillin") {
		t.Errorf("expected Penicillin allergy in response, got: %s", text)
	}
}

func TestHIS_GetDiagnosis_EmptyID(t *testing.T) {
	bin := buildHIS(t)
	resp := sendHIS(t, bin, `{"jsonrpc":"2.0","id":13,"method":"tools/call","params":{"name":"get_diagnosis","arguments":{"patient_id":""}}}`)
	if resp.Error == nil {
		t.Fatal("expected error for empty patient_id")
	}
	if resp.Error.Code != -32602 {
		t.Errorf("error code = %d, want -32602", resp.Error.Code)
	}
}
