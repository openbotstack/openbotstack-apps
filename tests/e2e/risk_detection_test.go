package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/openbotstack/openbotstack-core/assistant"
	"github.com/openbotstack/openbotstack-core/planner"
	"github.com/stretchr/testify/require"
)

func TestE2E_RiskDetection(t *testing.T) {
	ar := setupRuntime(t)
	ctx := context.Background()

	pCtx := &planner.PlannerContext{
		AssistantID: "test-icu-assistant",
		UserRequest: "Which ICU patients are critical? Please check P001.",
		Soul:        assistant.DefaultSoul(),
		Skills: []planner.SkillDescriptor{
			{ID: "ehr.query_vitals", Name: "Query Vitals", Description: "Query vitals for a patient"},
			{ID: "ehr.query_labs", Name: "Query Labs", Description: "Query labs for a patient"},
			{ID: "analytics.risk_score", Name: "Risk Score", Description: "Calculate risk from vitals and labs"},
		},
	}

	plan, err := ar.Planner.Plan(ctx, pCtx)
	require.NoError(t, err, "Planner should succeed")
	require.NotEmpty(t, plan.Steps, "Plan should have steps")

	// Usually it will pick vitals, labs, then risk score. Let's just execute all steps LLM planned.
	results, err := executePlanLocally(ctx, ar.ToolRunner.(*LocalToolRunner), plan)
	require.NoError(t, err, "Tool execution should succeed")
	require.NotEmpty(t, results)

	resultsJSON, _ := json.MarshalIndent(results, "", "  ")
	finalPrompt := "User wanted to know if patient P001 is critical. Tool data:\n" + string(resultsJSON) + "\n\nPlease write the final response to the user."

	var finalResponse string
	stream, err := ar.ModelProvider.Stream(finalPrompt)
	require.NoError(t, err)

	for token := range stream {
		finalResponse += token
	}

	require.NotEmpty(t, finalResponse, "Final response should not be empty")
	t.Logf("Final Response for Risk Detection:\n%s", finalResponse)
}
