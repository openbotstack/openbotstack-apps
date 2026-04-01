package workflows

import (
	"fmt"
	"time"

	"github.com/openbotstack/openbotstack-core/execution"
)

const shiftHandoverTimeout = 60 * time.Second

// ShiftHandoverWorkflow generates a nursing shift handover report
// for all patients in a given unit.
//
// Steps:
//  1. Query all patients in the unit           (tool: ehr.query_patient)
//  2. For a target patient, fetch vitals       (tool: ehr.query_vitals)
//  3. Fetch lab results                        (tool: ehr.query_labs)
//  4. Calculate risk score                     (tool: analytics.risk_score)
//  5. Generate SBAR handover                   (skill: nursing/generate_sbar)
type ShiftHandoverWorkflow struct{}

// ID implements Workflow.
func (w *ShiftHandoverWorkflow) ID() string { return "workflows/shift_handover" }

// Name implements Workflow.
func (w *ShiftHandoverWorkflow) Name() string { return "Nursing Shift Handover" }

// Description implements Workflow.
func (w *ShiftHandoverWorkflow) Description() string {
	return "Generate a comprehensive shift handover report for all patients in a nursing unit, " +
		"including vitals, labs, risk scores, and SBAR communications"
}

// Timeout implements Workflow.
func (w *ShiftHandoverWorkflow) Timeout() time.Duration { return shiftHandoverTimeout }

// Validate implements Workflow.
func (w *ShiftHandoverWorkflow) Validate() error {
	return nil
}

// Steps implements Workflow.
func (w *ShiftHandoverWorkflow) Steps(input map[string]any) ([]execution.ExecutionStep, error) {
	unit, _ := input["unit"].(string)
	if unit == "" {
		return nil, fmt.Errorf("input 'unit' is required for shift handover")
	}

	patientID, _ := input["patient_id"].(string)
	if patientID == "" {
		return nil, fmt.Errorf("input 'patient_id' is required for shift handover")
	}

	return []execution.ExecutionStep{
		{
			Name: "ehr.query_patient",
			Type: execution.StepTypeTool,
			Arguments: map[string]any{
				"unit": unit,
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
			Name: "nursing/generate_sbar",
			Type: execution.StepTypeSkill,
			Arguments: map[string]any{
				"patient_id": patientID,
			},
		},
	}, nil
}
