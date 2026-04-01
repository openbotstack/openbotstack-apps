package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/openbotstack/openbotstack-core/assistant"
	"github.com/openbotstack/openbotstack-core/planner"
)

func TestSoulInjection(t *testing.T) {
	ctx := context.Background()

	// 1. Create a mock LLM that captures the generated request
	mockLLM := &MockModelProvider{
		ResponseSequence: []string{
			`{
			  "assistant_id": "test-assistant",
			  "steps": []
			}`,
		},
	}
	router := &MockRouter{Provider: mockLLM}
	execPlanner := planner.NewLLMPlanner(router, nil)

	// 2. Define a custom soul with specific system prompt, identity, and personality
	customSoul := assistant.AssistantSoul{
		SystemPrompt: "You are an expert ICU Nursing Assistant.",
		Personality:  "Empathetic, strict, highly precise",
		Instructions: "- Always verify patient ID\n- Do not assume missing data is normal.",
	}

	pCtx := &planner.PlannerContext{
		UserRequest: "Check recent vitals.",
		Soul:        customSoul,
		Skills: []planner.SkillDescriptor{
			{ID: "ehr.query_vitals", Name: "Query Vitals"},
		},
	}

	// 3. Run Planner
	_, err := execPlanner.Plan(ctx, pCtx)
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}

	// 4. Verify that the RequestHistory captured the System Prompt correctly
	if len(mockLLM.RequestHistory) != 1 {
		t.Fatalf("Expected 1 LLM request, got %d", len(mockLLM.RequestHistory))
	}

	req := mockLLM.RequestHistory[0]
	if len(req.Messages) < 2 {
		t.Fatalf("Expected at least 2 messages (system + user), got %d", len(req.Messages))
	}

	// Verify System role has the core SystemPrompt
	sysMsg := req.Messages[0]
	if sysMsg.Role != "system" {
		t.Errorf("Expected first message to be role 'system', got '%s'", sysMsg.Role)
	}
	if sysMsg.Content != customSoul.SystemPrompt {
		t.Errorf("Expected system prompt content to be '%s', got '%s'", customSoul.SystemPrompt, sysMsg.Content)
	}

	// Verify User role dynamically incorporates Personality and Instructions
	userMsg := req.Messages[1]
	if !strings.Contains(userMsg.Content, customSoul.Personality) {
		t.Errorf("Expected user prompt to include personality '%s', but it didn't. Content: %s", customSoul.Personality, userMsg.Content)
	}
	if !strings.Contains(userMsg.Content, customSoul.Instructions) {
		t.Errorf("Expected user prompt to include instructions '%s', but it didn't. Content: %s", customSoul.Instructions, userMsg.Content)
	}
}
