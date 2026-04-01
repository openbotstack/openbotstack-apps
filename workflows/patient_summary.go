package workflows

import (
	"fmt"
	"time"

	"github.com/openbotstack/openbotstack-core/execution"
)

const patientSummaryTimeout = 30 * time.Second

// PatientSummaryWorkflow generates a clinical summary for a single patient.
//
// Steps:
//  1. Fetch patient demographics           (tool: ehr.query_patient)
//  2. Fetch current vitals                 (tool: ehr.query_vitals)
//  3. Fetch lab results                    (tool: ehr.query_labs)
//  4. Calculate risk score                 (tool: analytics.risk_score)
//  5. Summarize patient status             (skill: nursing/summarize_status)
type PatientSummaryWorkflow struct{}

// ID implements Workflow.
func (w *PatientSummaryWorkflow) ID() string { return "workflows/patient_summary" }

// Name implements Workflow.
func (w *PatientSummaryWorkflow) Name() string { return "Patient Summary" }

// Description implements Workflow.
func (w *PatientSummaryWorkflow) Description() string {
	return "Generate a comprehensive clinical summary for a single patient including " +
		"demographics, vitals, labs, risk score, and status assessment"
}

// Timeout implements Workflow.
func (w *PatientSummaryWorkflow) Timeout() time.Duration { return patientSummaryTimeout }

// Validate implements Workflow.
func (w *PatientSummaryWorkflow) Validate() error {
	return nil
}

// Steps implements Workflow.
func (w *PatientSummaryWorkflow) Steps(input map[string]any) ([]execution.ExecutionStep, error) {
	patientID, _ := input["patient_id"].(string)
	if patientID == "" {
		return nil, fmt.Errorf("input 'patient_id' is required for patient summary")
	}

	return []execution.ExecutionStep{
		{
			Name: "ehr.query_patient",
			Type: execution.StepTypeTool,
			Arguments: map[string]any{
				"patient_id": patientID,
			},
		},
		{
			Name: "ehr.query_vitals",
			Type: execution.StepTypeTool,
			Arguments: map[string]any{
				"patient_id": patientID,
			},
		},
		{
			Name: "ehr.query_labs",
			Type: execution.StepTypeTool,
			Arguments: map[string]any{
				"patient_id": patientID,
			},
		},
		{
			Name: "analytics.risk_score",
			Type: execution.StepTypeTool,
			Arguments: map[string]any{
				"patient_id": patientID,
			},
		},
		{
			Name: "nursing/summarize_status",
			Type: execution.StepTypeSkill,
			Arguments: map[string]any{
				"patient_id": patientID,
			},
		},
	}, nil
}
