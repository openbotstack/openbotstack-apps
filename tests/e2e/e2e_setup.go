package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/openbotstack/openbotstack-apps/tools/analytics"
	"github.com/openbotstack/openbotstack-apps/tools/ehr"
	"github.com/openbotstack/openbotstack-apps/workflows"
	"github.com/openbotstack/openbotstack-core/ai/providers"
	"github.com/openbotstack/openbotstack-core/control/skills"
	"github.com/openbotstack/openbotstack-core/execution"
	"github.com/openbotstack/openbotstack-core/planner"
	"github.com/openbotstack/openbotstack-runtime/runtime"
	"github.com/openbotstack/openbotstack-runtime/runtime/llm"
	"github.com/openbotstack/openbotstack-runtime/runtime/memory"
)

// LocalToolRunner executes locally registered tools, bypassing network registries.
type LocalToolRunner struct{}

func (r *LocalToolRunner) Execute(ctx context.Context, name string, input map[string]any, ec *execution.ExecutionContext) (*execution.StepResult, error) {
	var res any
	var err error

	switch name {
	case "ehr.query_patient":
		res, err = (&ehr.QueryPatientTool{}).Execute(ctx, input)
	case "ehr.query_vitals":
		res, err = (&ehr.QueryVitalsTool{}).Execute(ctx, input)
	case "ehr.query_labs":
		res, err = (&ehr.QueryLabsTool{}).Execute(ctx, input)
	case "analytics.risk_score":
		res, err = (&analytics.RiskScoreTool{}).Execute(ctx, input)
	case "nursing/generate_sbar":
		res = map[string]string{"sbar": "Shift handover SBAR generated for " + input["patient_id"].(string)}
	case "nursing/summarize_status":
		res = map[string]string{"summary": "Patient status summarized for " + input["patient_id"].(string)}
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}

	if err != nil {
		return &execution.StepResult{StepName: name, Error: err}, err
	}

	return &execution.StepResult{
		StepName: name,
		Output:   res,
		Type:     string(execution.StepTypeTool),
	}, nil
}

type MockRouter struct {
	Provider providers.ModelProvider
}

func (m *MockRouter) Route(reqs []skills.CapabilityType, constraints skills.ModelConstraints) (providers.ModelProvider, error) {
	return m.Provider, nil
}

func (m *MockRouter) Register(p providers.ModelProvider) error { return nil }

func (m *MockRouter) List() []string { return []string{"mock"} }

type ProviderWrapper struct {
	llm.ModelProvider
	apiKey string
}

func (w *ProviderWrapper) ID() string { return "openai/gpt-4o" }
func (w *ProviderWrapper) Capabilities() []skills.CapabilityType {
	return []skills.CapabilityType{skills.CapTextGeneration}
}
func (w *ProviderWrapper) Generate(ctx context.Context, req skills.GenerateRequest) (*skills.GenerateResponse, error) {
	prompt := ""
	for _, m := range req.Messages {
		prompt += fmt.Sprintf("%s: %s\n", m.Role, m.Content)
	}
	res, err := w.ModelProvider.Generate(prompt)
	if err != nil {
		return nil, err
	}
	return &skills.GenerateResponse{Content: res}, nil
}
func (w *ProviderWrapper) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	return nil, nil
}

func setupRuntime(t *testing.T) *runtime.AssistantRuntime {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping real E2E test")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping real E2E test")
	}

	// In a real ICU app, we'd use pgxpool. NewPostgresMemoryManager takes *pgxpool.Pool.
	// For testing, we just mock or skip if no real DB.
	memManager := memory.NewPostgresMemoryManager(nil)

	baseProvider := llm.NewOpenAIProvider(apiKey)
	provider := &ProviderWrapper{ModelProvider: baseProvider, apiKey: apiKey}

	router := &MockRouter{Provider: provider}
	plannerImpl := planner.NewLLMPlanner(router, nil)

	cfg := runtime.Config{
		SoulPath:      "soul.md",
		MemoryManager: memManager,
		Planner:       plannerImpl,
		ToolRunner:    &LocalToolRunner{},
		ModelProvider: baseProvider,
	}
	ar, err := runtime.NewAssistantRuntime(cfg)
	if err != nil {
		t.Fatalf("failed to init assistant runtime: %v", err)
	}
	return ar
}

func executePlanLocally(ctx context.Context, tr *LocalToolRunner, plan *execution.ExecutionPlan) ([]*execution.StepResult, error) {
	var results []*execution.StepResult
	for _, step := range plan.Steps {
		// If it's a workflow, we build the sub-plan and execute it recursively
		if step.Name == "workflow.shift_handover" {
			subPlan, _ := workflows.BuildPlan(&workflows.ShiftHandoverWorkflow{}, step.Arguments)
			subResults, err := executePlanLocally(ctx, tr, subPlan)
			if err != nil {
				return nil, err
			}
			results = append(results, subResults...)
			continue
		}
		if step.Name == "workflow.patient_summary" {
			subPlan, _ := workflows.BuildPlan(&workflows.PatientSummaryWorkflow{}, step.Arguments)
			subResults, err := executePlanLocally(ctx, tr, subPlan)
			if err != nil {
				return nil, err
			}
			results = append(results, subResults...)
			continue
		}

		res, err := tr.Execute(ctx, step.Name, step.Arguments, nil)
		if err != nil {
			return nil, err
		}
		results = append(results, res)
	}
	return results, nil
}
