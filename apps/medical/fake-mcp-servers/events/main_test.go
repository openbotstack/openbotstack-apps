package main

import (
	"encoding/json"
	"os/exec"
	"strings"
	"testing"
)

func buildEvents(t *testing.T) string {
	t.Helper()
	bin := t.TempDir() + "/events-server"
	out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput()
	if err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	return bin
}

func sendEvents(t *testing.T, bin, input string) jsonRPCResponse {
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

func TestEvents_Initialize(t *testing.T) {
	bin := buildEvents(t)
	resp := sendEvents(t, bin, `{"jsonrpc":"2.0","id":1,"method":"initialize"}`)
	if resp.Error != nil {
		t.Fatalf("initialize error: %s", resp.Error.Message)
	}
	result := resp.Result.(map[string]any)
	info := result["serverInfo"].(map[string]any)
	if info["name"] != "fake-events" {
		t.Errorf("server name = %v, want fake-events", info["name"])
	}
}

func TestEvents_ToolsList(t *testing.T) {
	bin := buildEvents(t)
	resp := sendEvents(t, bin, `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`)
	result := resp.Result.(map[string]any)
	tools := result["tools"].([]any)
	if len(tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(tools))
	}
	toolNames := make(map[string]bool)
	for _, tool := range tools {
		tm := tool.(map[string]any)
		toolNames[tm["name"].(string)] = true
	}
	for _, name := range []string{"get_recent_events", "get_current_status"} {
		if !toolNames[name] {
			t.Errorf("missing tool: %s", name)
		}
	}
}

func TestEvents_GetRecentEvents(t *testing.T) {
	bin := buildEvents(t)
	resp := sendEvents(t, bin, `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_recent_events","arguments":{"patient_id":"P001"}}}`)
	if resp.Error != nil {
		t.Fatalf("get_recent_events error: %s", resp.Error.Message)
	}
	content := resp.Result.(map[string]any)["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)
	if !strings.Contains(text, "P001") {
		t.Errorf("expected P001 in response, got: %s", text)
	}
	if !strings.Contains(text, "Vancomycin") {
		t.Errorf("expected event data in response, got: %s", text)
	}
}

func TestEvents_GetRecentEvents_NotFound(t *testing.T) {
	bin := buildEvents(t)
	resp := sendEvents(t, bin, `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_recent_events","arguments":{"patient_id":"P999"}}}`)
	if resp.Error == nil {
		t.Fatal("expected error for unknown patient")
	}
	if resp.Error.Code != -32001 {
		t.Errorf("error code = %d, want -32001", resp.Error.Code)
	}
}

func TestEvents_GetCurrentStatus(t *testing.T) {
	bin := buildEvents(t)
	resp := sendEvents(t, bin, `{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"get_current_status","arguments":{"patient_id":"P003"}}}`)
	if resp.Error != nil {
		t.Fatalf("get_current_status error: %s", resp.Error.Message)
	}
	content := resp.Result.(map[string]any)["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)
	if !strings.Contains(text, "P003") {
		t.Errorf("expected P003 in response, got: %s", text)
	}
	if !strings.Contains(text, "critical") {
		t.Errorf("expected critical status for P003 (has AMI), got: %s", text)
	}
}

func TestEvents_GetCurrentStatus_Stable(t *testing.T) {
	bin := buildEvents(t)
	resp := sendEvents(t, bin, `{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"get_current_status","arguments":{"patient_id":"P002"}}}`)
	if resp.Error != nil {
		t.Fatalf("get_current_status error: %s", resp.Error.Message)
	}
	content := resp.Result.(map[string]any)["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)
	if !strings.Contains(text, "P002") {
		t.Errorf("expected P002 in response, got: %s", text)
	}
}
