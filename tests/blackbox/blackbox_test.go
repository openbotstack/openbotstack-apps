package blackbox

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/openbotstack/openbotstack-apps/tools/analytics"
	"github.com/openbotstack/openbotstack-apps/tools/ehr"
	"github.com/openbotstack/openbotstack-core/ai/providers"
	"github.com/openbotstack/openbotstack-core/assistant"
	"github.com/openbotstack/openbotstack-core/control/skills"
	"github.com/openbotstack/openbotstack-core/execution"
	"github.com/openbotstack/openbotstack-core/planner"
	"github.com/openbotstack/openbotstack-runtime/runtime"
	"github.com/openbotstack/openbotstack-runtime/runtime/llm"
	"github.com/openbotstack/openbotstack-runtime/runtime/memory"
	"github.com/stretchr/testify/require"
)

type FailingToolRunner struct{}

func (r *FailingToolRunner) Execute(ctx context.Context, name string, input map[string]any, ec *execution.ExecutionContext) (*execution.StepResult, error) {
	return &execution.StepResult{
		StepName: name,
		Error:    fmt.Errorf("simulated tool failure for %s", name),
	}, fmt.Errorf("simulated tool failure for %s", name)
}

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
		res = map[string]string{"sbar": "Shift handover SBAR generated"}
	case "nursing/summarize_status":
		res = map[string]string{"summary": "Patient status summarized"}
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

func setupRuntime(t *testing.T, toolRunner toolrunnerIfc) *runtime.AssistantRuntime {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	// Mock DB components for blackbox tests
	memManager := memory.NewPostgresMemoryManager(nil)

	baseProvider := llm.NewOpenAIProvider(apiKey)
	provider := &ProviderWrapper{ModelProvider: baseProvider}
	router := &MockRouter{Provider: provider}
	plannerImpl := planner.NewLLMPlanner(router, nil)

	cfg := runtime.Config{
		SoulPath:      "soul.md",
		MemoryManager: memManager,
		Planner:       plannerImpl,
		ToolRunner:    toolRunner,
		ModelProvider: baseProvider,
	}
	ar, err := runtime.NewAssistantRuntime(cfg)
	if err != nil {
		t.Fatalf("failed to init assistant runtime: %v", err)
	}
	return ar
}

type toolrunnerIfc interface {
	Execute(ctx context.Context, name string, input map[string]any, ec *execution.ExecutionContext) (*execution.StepResult, error)
}

// TestBlackBox_NaturalLanguage exercises the planner with unstructured natural language prompts
func TestBlackBox_NaturalLanguage(t *testing.T) {
	ar := setupRuntime(t, &LocalToolRunner{})
	ctx := context.Background()

	prompts := []string{
		"Who is unstable?",
		"Which patient needs attention?",
		"Prepare ICU handover",
	}

	availableSkills := []planner.SkillDescriptor{
		{ID: "workflow.shift_handover", Name: "Shift Handover", Description: "Generates handover for a patient in a unit"},
		{ID: "ehr.query_vitals", Name: "Query Vitals", Description: "Query vitals for a patient"},
		{ID: "ehr.query_labs", Name: "Query Labs", Description: "Query labs for a patient"},
		{ID: "analytics.risk_score", Name: "Risk Score", Description: "Calculate risk from vitals and labs"},
	}

	for _, prompt := range prompts {
		t.Run(prompt, func(t *testing.T) {
			pCtx := &planner.PlannerContext{
				AssistantID: "test-blackbox",
				UserRequest: prompt,
				Soul:        assistant.DefaultSoul(),
				Skills:      availableSkills,
			}

			plan, err := ar.Planner.Plan(ctx, pCtx)
			require.NoError(t, err)
			require.NotEmpty(t, plan.Steps, "Planner should formulate steps even for vague prompts")
			t.Logf("Prompt: %s => Plan steps: %d", prompt, len(plan.Steps))
		})
	}
}

// TestBlackBox_Failure simulates tool failure and LLM integration failure
func TestBlackBox_Failure(t *testing.T) {
	// 1. Setup with a failing tool runner to test step error capture
	ar := setupRuntime(t, &FailingToolRunner{})
	ctx := context.Background()

	pCtx := &planner.PlannerContext{
		AssistantID: "test-fail",
		UserRequest: "Run tool ehr.query_patient",
		Soul:        assistant.DefaultSoul(),
		Skills: []planner.SkillDescriptor{
			{ID: "ehr.query_patient", Name: "Query Patient", Description: "queries a patient"},
		},
	}

	plan, err := ar.Planner.Plan(ctx, pCtx)
	require.NoError(t, err)

	// Execute it through the failing runner
	require.NotEmpty(t, plan.Steps)
	res, err := ar.ToolRunner.(*FailingToolRunner).Execute(ctx, plan.Steps[0].Name, plan.Steps[0].Arguments, nil)
	
	require.Error(t, err, "Should capture tool failure")
	require.Contains(t, err.Error(), "simulated tool failure")
	require.NotNil(t, res)
	require.Error(t, res.Error, "StepResult should contain the error")

	// 2. Validate broken LLM configuration fails gracefully during planning
	brokengptProv := &ProviderWrapper{ModelProvider: llm.NewOpenAIProvider("sk-invalid-key")}
	brokenRouter := &MockRouter{Provider: brokengptProv}
	brokenPlanner := planner.NewLLMPlanner(brokenRouter, nil)

	_, err = brokenPlanner.Plan(ctx, pCtx)
	require.Error(t, err, "Planning with bad auth key should fail gracefully")
}
