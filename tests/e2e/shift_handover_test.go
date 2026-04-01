package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/openbotstack/openbotstack-core/assistant"
	"github.com/openbotstack/openbotstack-core/planner"
	"github.com/stretchr/testify/require"
)

func TestE2E_ShiftHandover(t *testing.T) {
	ar := setupRuntime(t)
	ctx := context.Background()

	pCtx := &planner.PlannerContext{
		AssistantID: "test-icu-assistant",
		UserRequest: "Generate ICU shift handover for patient P001",
		Soul:        assistant.DefaultSoul(),
		Skills: []planner.SkillDescriptor{
			{ID: "workflow.shift_handover", Name: "Shift Handover", Description: "Generates handover for a patient in a unit"},
		},
	}

	// 1. Generate Execution Plan
	plan, err := ar.Planner.Plan(ctx, pCtx)
	require.NoError(t, err, "Planner should succeed")
	require.NotEmpty(t, plan.Steps, "Plan should have steps")
	
	// Verify it picked the workflow
	require.Equal(t, "workflow.shift_handover", plan.Steps[0].Name)

	// 2. Execute Tools
	results, err := executePlanLocally(ctx, ar.ToolRunner.(*LocalToolRunner), plan)
	require.NoError(t, err, "Tool execution should succeed")
	require.Len(t, results, 5, "Shift handover should expand to 5 tool steps")

	// 3. Final Output via LLM
	// In a complete loop, the tool outputs form the context for the final generation
	resultsJSON, _ := json.Marshal(results)
	finalPrompt := "User wanted a shift handover. Here is the tool data gathered:\n" + string(resultsJSON) + "\n\nPlease write the final response."

	var finalResponse string
	stream, err := ar.ModelProvider.Stream(finalPrompt)
	require.NoError(t, err)

	for token := range stream {
		finalResponse += token
	}

	require.NotEmpty(t, finalResponse, "Final response should not be empty")
	t.Logf("Final Response Configuration:\n%s", finalResponse)
}
