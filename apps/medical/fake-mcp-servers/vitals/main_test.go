package main

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

func buildVitals(t *testing.T) string {
	t.Helper()
	bin := t.TempDir() + "/vitals-server"
	out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput()
	if err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	return bin
}

func sendVitals(t *testing.T, bin, input string) jsonRPCResponse {
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

func TestVitals_Initialize(t *testing.T) {
	bin := buildVitals(t)
	resp := sendVitals(t, bin, `{"jsonrpc":"2.0","id":1,"method":"initialize"}`)
	if resp.Error != nil {
		t.Fatalf("error: %s", resp.Error.Message)
	}
	info := resp.Result.(map[string]any)["serverInfo"].(map[string]any)
	if info["name"] != "fake-vitals" {
		t.Errorf("name = %v", info["name"])
	}
}

func TestVitals_ToolsList(t *testing.T) {
	bin := buildVitals(t)
	resp := sendVitals(t, bin, `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)
	tools := resp.Result.(map[string]any)["tools"].([]any)
	if len(tools) != 4 {
		t.Fatalf("expected 4 tools, got %d", len(tools))
	}
}

func TestVitals_QueryVitals(t *testing.T) {
	bin := buildVitals(t)
	resp := sendVitals(t, bin, `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"query_vitals","arguments":{"patient_id":"P004"}}}`)
	if resp.Error != nil {
		t.Fatalf("error: %s", resp.Error.Message)
	}
	text := extractText(t, resp)
	var vitals map[string]any
	json.Unmarshal([]byte(text), &vitals)
	if vitals["on_ventilator"] != true {
		t.Error("P004 should be on ventilator")
	}
}

func TestVitals_Alerts(t *testing.T) {
	bin := buildVitals(t)
	resp := sendVitals(t, bin, `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_alerts","arguments":{"patient_id":"P001"}}}`)
	if resp.Error != nil {
		t.Fatalf("error: %s", resp.Error.Message)
	}
	text := extractText(t, resp)
	var result map[string]any
	json.Unmarshal([]byte(text), &result)
	if len(result["alerts"].([]any)) == 0 {
		t.Error("P001 should have alerts")
	}
}

func TestVitals_Trend(t *testing.T) {
	bin := buildVitals(t)
	resp := sendVitals(t, bin, `{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"get_vital_trend","arguments":{"patient_id":"P001","vital_sign":"heart_rate","hours":12}}}`)
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

func TestVitals_GetVitals(t *testing.T) {
	bin := buildVitals(t)
	resp := sendVitals(t, bin, `{"jsonrpc":"2.0","id":10,"method":"tools/call","params":{"name":"get_vitals","arguments":{"patient_id":"P001"}}}`)
	if resp.Error != nil {
		t.Fatalf("get_vitals error: %s", resp.Error.Message)
	}
	content := resp.Result.(map[string]any)["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)
	if !strings.Contains(text, "P001") {
		t.Errorf("expected P001 in response, got: %s", text)
	}
	if !strings.Contains(text, "heart_rate") {
		t.Errorf("expected heart_rate in response, got: %s", text)
	}
}
