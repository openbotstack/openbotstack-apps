package main

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

func buildLIS(t *testing.T) string {
	t.Helper()
	bin := t.TempDir() + "/lis-server"
	out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput()
	if err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	return bin
}

func sendLIS(t *testing.T, bin, input string) jsonRPCResponse {
	t.Helper()
	cmd := exec.Command(bin)
	cmd.Stdin = strings.NewReader(input + "\n")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run: %v\n%s", err, out)
	}
	var resp jsonRPCResponse
	json.Unmarshal([]byte(strings.TrimSpace(string(out))), &resp)
	return resp
}

func TestLIS_Initialize(t *testing.T) {
	bin := buildLIS(t)
	resp := sendLIS(t, bin, `{"jsonrpc":"2.0","id":1,"method":"initialize"}`)
	if resp.Error != nil {
		t.Fatalf("error: %s", resp.Error.Message)
	}
	info := resp.Result.(map[string]any)["serverInfo"].(map[string]any)
	if info["name"] != "fake-lis" {
		t.Errorf("name = %v", info["name"])
	}
}

func TestLIS_ToolsList(t *testing.T) {
	bin := buildLIS(t)
	resp := sendLIS(t, bin, `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)
	tools := resp.Result.(map[string]any)["tools"].([]any)
	if len(tools) != 4 {
		t.Fatalf("expected 4 tools, got %d", len(tools))
	}
}

func TestLIS_QueryLabs(t *testing.T) {
	bin := buildLIS(t)
	resp := sendLIS(t, bin, `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"query_labs","arguments":{"patient_id":"P001"}}}`)
	if resp.Error != nil {
		t.Fatalf("error: %s", resp.Error.Message)
	}
	text := extractText(t, resp)
	var panel map[string]any
	json.Unmarshal([]byte(text), &panel)
	if len(panel["results"].([]any)) == 0 {
		t.Fatal("no lab results")
	}
}

func TestLIS_CriticalValues(t *testing.T) {
	bin := buildLIS(t)
	resp := sendLIS(t, bin, `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_critical_values","arguments":{"patient_id":"P001"}}}`)
	if resp.Error != nil {
		t.Fatalf("error: %s", resp.Error.Message)
	}
	text := extractText(t, resp)
	var result map[string]any
	json.Unmarshal([]byte(text), &result)
	if len(result["critical_values"].([]any)) == 0 {
		t.Error("P001 should have critical values")
	}
}

func TestLIS_Trend(t *testing.T) {
	bin := buildLIS(t)
	resp := sendLIS(t, bin, `{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"get_trend","arguments":{"patient_id":"P001","test_name":"CRP","hours":24}}}`)
	if resp.Error != nil {
		t.Fatalf("error: %s", resp.Error.Message)
	}
	text := extractText(t, resp)
	var result map[string]any
	json.Unmarshal([]byte(text), &result)
	if len(result["trend"].([]any)) == 0 {
		t.Error("expected trend data")
	}
}

func extractText(t *testing.T, resp jsonRPCResponse) string {
	t.Helper()
	content := resp.Result.(map[string]any)["content"].([]any)
	return content[0].(map[string]any)["text"].(string)
}

func TestLIS_GetLabResults(t *testing.T) {
	bin := buildLIS(t)
	resp := sendLIS(t, bin, `{"jsonrpc":"2.0","id":10,"method":"tools/call","params":{"name":"get_lab_results","arguments":{"patient_id":"P001"}}}`)
	if resp.Error != nil {
		t.Fatalf("get_lab_results error: %s", resp.Error.Message)
	}
	content := resp.Result.(map[string]any)["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)
	if !strings.Contains(text, "WBC") {
		t.Errorf("expected WBC in response, got: %s", text)
	}
	if !strings.Contains(text, "CRP") {
		t.Errorf("expected CRP in response, got: %s", text)
	}
}

func TestLIS_GetLabResults_EmptyID(t *testing.T) {
	bin := buildLIS(t)
	resp := sendLIS(t, bin, `{"jsonrpc":"2.0","id":11,"method":"tools/call","params":{"name":"get_lab_results","arguments":{"patient_id":""}}}`)
	if resp.Error == nil {
		t.Fatal("expected error for empty patient_id")
	}
	if resp.Error.Code != -32602 {
		t.Errorf("error code = %d, want -32602", resp.Error.Code)
	}
}
