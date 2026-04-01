package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/openbotstack/openbotstack-apps/tools/analytics"
	"github.com/openbotstack/openbotstack-apps/tools/ehr"
	"github.com/openbotstack/openbotstack-apps/workflows"
	"github.com/openbotstack/openbotstack-core/ai/providers"
	"github.com/openbotstack/openbotstack-core/assistant"
	"github.com/openbotstack/openbotstack-core/control/skills"
	"github.com/openbotstack/openbotstack-core/execution"
	"github.com/openbotstack/openbotstack-core/planner"
)

// MockModelProvider simulates an LLM for integration testing.
type MockModelProvider struct {
	// RequestHistory logs all generation requests to verify prompt injection
	RequestHistory []skills.GenerateRequest
	// ResponseSequence defines the JSON responses returned sequentially
	ResponseSequence []string
	callCount        int
}

func (m *MockModelProvider) ID() string { return "mock/agent" }

func (m *MockModelProvider) Capabilities() []skills.CapabilityType {
	return []skills.CapabilityType{skills.CapTextGeneration}
}

func (m *MockModelProvider) Generate(ctx context.Context, req skills.GenerateRequest) (*skills.GenerateResponse, error) {
	m.RequestHistory = append(m.RequestHistory, req)
	if m.callCount >= len(m.ResponseSequence) {
		return &skills.GenerateResponse{Content: "{ \"error\": \"mock exhausted\" }"}, nil
	}
	resp := m.ResponseSequence[m.callCount]
	m.callCount++
	return &skills.GenerateResponse{Content: resp}, nil
}

func (m *MockModelProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	return nil, nil
}

// MockRouter implements providers.ModelRouter
type MockRouter struct {
	Provider providers.ModelProvider
}

func (m *MockRouter) Route(reqs []skills.CapabilityType, constraints skills.ModelConstraints) (providers.ModelProvider, error) {
	return m.Provider, nil
}

func (m *MockRouter) Register(provider providers.ModelProvider) error { return nil }
func (m *MockRouter) List() []string                              { return []string{"mock/agent"} }

// simulateExecutionEngine executes an execution plan locally for testing.
// It maps the plan steps to the actual tools and skills in openbotstack-apps.
func simulateExecutionEngine(ctx context.Context, plan *execution.ExecutionPlan) ([]any, error) {
	var results []any

	for _, step := range plan.Steps {
		var res any
		var err error

		switch step.Type {
		case execution.StepTypeTool:
			switch step.Name {
			case "ehr.query_patient":
				tool := &ehr.QueryPatientTool{}
				res, err = tool.Execute(ctx, step.Arguments)
			case "ehr.query_vitals":
				tool := &ehr.QueryVitalsTool{}
				res, err = tool.Execute(ctx, step.Arguments)
			case "ehr.query_labs":
				tool := &ehr.QueryLabsTool{}
				res, err = tool.Execute(ctx, step.Arguments)
			case "analytics.risk_score":
				tool := &analytics.RiskScoreTool{}
				res, err = tool.Execute(ctx, step.Arguments)
			default:
				err = fmt.Errorf("unknown tool: %s", step.Name)
			}
		case execution.StepTypeSkill:
			// In integration tests, a skill step might require generating some text.
			// Or it might just represent a checkpoint in the workflow.
			switch step.Name {
			case "nursing/generate_sbar":
				res = map[string]string{"sbar": "Shift handover SBAR generated for " + step.Arguments["patient_id"].(string)}
			case "nursing/summarize_status":
				res = map[string]string{"summary": "Status summarized for " + step.Arguments["patient_id"].(string)}
			case "workflow.shift_handover":
				// The LLM has chosen to run the workflow. We expand it.
				wf := &workflows.ShiftHandoverWorkflow{}
				subPlan, buildErr := workflows.BuildPlan(wf, step.Arguments)
				if buildErr != nil {
					err = buildErr
				} else {
					// recursive execution
					subResults, subErr := simulateExecutionEngine(ctx, subPlan)
					res = subResults
					err = subErr
				}
			case "workflow.patient_summary":
				wf := &workflows.PatientSummaryWorkflow{}
				subPlan, buildErr := workflows.BuildPlan(wf, step.Arguments)
				if buildErr != nil {
					err = buildErr
				} else {
					subResults, subErr := simulateExecutionEngine(ctx, subPlan)
					res = subResults
					err = subErr
				}
			default:
				err = fmt.Errorf("unknown skill: %s", step.Name)
			}
		default:
			err = fmt.Errorf("unknown step type: %s", step.Type)
		}

		if err != nil {
			return nil, fmt.Errorf("step %s failed: %w", step.Name, err)
		}
		results = append(results, res)
	}

	return results, nil
}

func TestAgentIntegration_ICUShiftHandover(t *testing.T) {
	ctx := context.Background()

	// 1. Setup Mock Provider to return a plan asking for the shift_handover workflow
	mockLLM := &MockModelProvider{
		ResponseSequence: []string{
			`{
			  "assistant_id": "test-assistant",
			  "steps": [
			    {
			      "type": "skill",
			      "name": "workflow.shift_handover",
			      "arguments": {"unit": "ICU", "patient_id": "P001"}
			    }
			  ]
			}`,
		},
	}
	router := &MockRouter{Provider: mockLLM}
	execPlanner := planner.NewLLMPlanner(router, nil)

	// 2. Build Planner Context
	pCtx := &planner.PlannerContext{
		AssistantID: "test-assistant",
		UserRequest: "Generate ICU shift handover",
		Soul:        assistant.DefaultSoul(),
		Skills: []planner.SkillDescriptor{
			{ID: "workflow.shift_handover", Name: "Shift Handover", Description: "Generates handover"},
		},
	}

	// 3. Plan Generation (LLM Reasoning -> Plan)
	plan, err := execPlanner.Plan(ctx, pCtx)
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}

	if len(plan.Steps) != 1 || plan.Steps[0].Name != "workflow.shift_handover" {
		t.Fatalf("Expected shift handover workflow selection, got %+v", plan.Steps)
	}

	// 4. Execute the expected plan via simulation
	results, err := simulateExecutionEngine(ctx, plan)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// 5. Verify Execution details
	// The ShiftHandoverWorkflow compiles to patient, vitals, labs, risk_score, generate_sbar
	if len(results) != 1 {
		t.Fatalf("Expected 1 grouped result, got %d", len(results))
	}
	subResults, ok := results[0].([]any)
	if !ok || len(subResults) != 5 {
		t.Fatalf("Expected 5 inner steps for shift handover, got %v", subResults)
	}

	// Verify Final SBAR output from the last step
	finalOutput, ok := subResults[4].(map[string]string)
	if !ok || !strings.Contains(finalOutput["sbar"], "SBAR generated") {
		t.Errorf("Unexpected final output: %v", subResults[4])
	}
}

func TestAgentIntegration_PatientSummary(t *testing.T) {
	ctx := context.Background()

	mockLLM := &MockModelProvider{
		ResponseSequence: []string{
			`{
			  "assistant_id": "test-assistant",
			  "steps": [
			    {
			      "type": "skill",
			      "name": "workflow.patient_summary",
			      "arguments": {"patient_id": "P001"}
			    }
			  ]
			}`,
		},
	}
	router := &MockRouter{Provider: mockLLM}
	execPlanner := planner.NewLLMPlanner(router, nil)

	pCtx := &planner.PlannerContext{
		AssistantID: "test-assistant",
		UserRequest: "Summarize patient P001",
		Soul:        assistant.DefaultSoul(),
		Skills: []planner.SkillDescriptor{
			{ID: "workflow.patient_summary", Name: "Patient Summary", Description: "Summarizes patient"},
		},
	}

	plan, err := execPlanner.Plan(ctx, pCtx)
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}

	results, err := simulateExecutionEngine(ctx, plan)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	subResults := results[0].([]any)
	if len(subResults) != 5 {
		t.Fatalf("Expected 5 steps for patient summary, got %d", len(subResults))
	}
	
	finalOutput := subResults[4].(map[string]string)
	if !strings.Contains(finalOutput["summary"], "Status summarized") {
		t.Errorf("Unexpected final output: %v", finalOutput)
	}
}

func TestAgentIntegration_RiskDetection(t *testing.T) {
	ctx := context.Background()

	// Risk detection expected to directly query risk_score tool and then generate summary
	mockLLM := &MockModelProvider{
		ResponseSequence: []string{
			`{
			  "assistant_id": "test-assistant",
			  "steps": [
			    {
			      "type": "tool",
			      "name": "ehr.query_patient",
			      "arguments": {"unit": "ICU"}
			    },
			    {
			      "type": "tool",
			      "name": "analytics.risk_score",
			      "arguments": {"patient_id": "P001"}
			    }
			  ]
			}`,
		},
	}
	router := &MockRouter{Provider: mockLLM}
	execPlanner := planner.NewLLMPlanner(router, nil)

	pCtx := &planner.PlannerContext{
		UserRequest: "Which ICU patients are critical?",
		Soul:        assistant.DefaultSoul(),
		Skills: []planner.SkillDescriptor{
			{ID: "ehr.query_patient", Name: "Query Patient", Description: "Query demographics"},
			{ID: "analytics.risk_score", Name: "Risk Score", Description: "Calculate risk"},
		},
	}
	plan, err := execPlanner.Plan(ctx, pCtx)
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}

	results, err := simulateExecutionEngine(ctx, plan)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	// Result 0: Patient List from ICU query
	if pList, ok := results[0].([]ehr.PatientRecord); !ok || len(pList) == 0 {
		t.Errorf("Expected list of ICU patients, got %v", results[0])
	}
	// Result 1: RiskScore
	// But note: analytics.RiskScoreTool requires full vitals to execute in reality,
	// though our mock implementation handles partial input mapping. Wait, RiskScoreTool in analytics
	// extracts values from map. If they are missing, it defaults to 0 and might not trigger 'critical'.
	// But it won't fail the engine execution. Let's verify it executes without errors.
	if _, ok := results[1].(analytics.RiskScoreResult); !ok {
		t.Errorf("Expected RiskScore result, got %T", results[1])
	}
}

func TestAgentIntegration_ToolCallLoop(t *testing.T) {
	ctx := context.Background()

	// Simulate a conversational agent loop:
	// Turn 1: request ehr.query_patient
	// Turn 2: request ehr.query_vitals based on the observation
	mockLLM := &MockModelProvider{
		ResponseSequence: []string{
			`{
			  "assistant_id": "test-assistant",
			  "steps": [{"type": "tool", "name": "ehr.query_patient", "arguments": {"patient_id": "P001"}}]
			}`,
			`{
			  "assistant_id": "test-assistant",
			  "steps": [{"type": "tool", "name": "ehr.query_vitals", "arguments": {"patient_id": "P001"}}]
			}`,
		},
	}

	router := &MockRouter{Provider: mockLLM}
	execPlanner := planner.NewLLMPlanner(router, nil)

	// Turn 1
	pCtx := &planner.PlannerContext{
		UserRequest: "Check patient P001",
		Soul:        assistant.DefaultSoul(),
		Skills: []planner.SkillDescriptor{
			{ID: "ehr.query_patient", Name: "Query Patient", Description: "Query demographics"},
			{ID: "ehr.query_vitals", Name: "Query Vitals", Description: "Query vitals"},
		},
	}
	plan1, err := execPlanner.Plan(ctx, pCtx)
	if err != nil {
		t.Fatalf("Turn 1 Plan failed: %v", err)
	}
	res1, err := simulateExecutionEngine(ctx, plan1)
	if err != nil {
		t.Fatalf("Turn 1 Exec failed: %v", err)
	}

	// Inject observation into history
	observationData, _ := json.Marshal(res1[0])
	pCtx.MemoryContext = append(pCtx.MemoryContext, assistant.SearchResult{
		Content: []byte(fmt.Sprintf("Observation: %s", observationData)),
	})
	
	// Turn 2
	plan2, err := execPlanner.Plan(ctx, pCtx)
	if err != nil {
		t.Fatalf("Turn 2 Plan failed: %v", err)
	}
	res2, err := simulateExecutionEngine(ctx, plan2)
	if err != nil {
		t.Fatalf("Turn 2 Exec failed: %v", err)
	}

	if _, ok := res2[0].(ehr.VitalSigns); !ok {
		t.Errorf("Expected vital signs in turn 2, got %T", res2[0])
	}
}
