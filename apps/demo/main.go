// Command demo is a CLI tool that demonstrates the openbotstack-apps
// application plane capabilities.
//
// It exercises tools, skills, and workflows against mock data
// without requiring a running runtime or LLM backend.
//
// Usage:
//
//	go run ./apps/demo
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/openbotstack/openbotstack-apps/tools/analytics"
	"github.com/openbotstack/openbotstack-apps/tools/ehr"
	"github.com/openbotstack/openbotstack-apps/workflows"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()

	fmt.Println("=== OpenBotStack Application Plane Demo ===")
	fmt.Println()

	// --- Tool layer demo ---
	fmt.Println("--- EHR: Query Patient P001 ---")
	patientTool := &ehr.QueryPatientTool{}
	patient, err := patientTool.Execute(ctx, map[string]any{"patient_id": "P001"})
	if err != nil {
		return fmt.Errorf("query patient: %w", err)
	}
	printJSON(patient)

	fmt.Println()
	fmt.Println("--- EHR: Query Vitals P001 ---")
	vitalsTool := &ehr.QueryVitalsTool{}
	vitals, err := vitalsTool.Execute(ctx, map[string]any{"patient_id": "P001"})
	if err != nil {
		return fmt.Errorf("query vitals: %w", err)
	}
	printJSON(vitals)

	fmt.Println()
	fmt.Println("--- EHR: Query Labs P001 ---")
	labsTool := &ehr.QueryLabsTool{}
	labs, err := labsTool.Execute(ctx, map[string]any{"patient_id": "P001"})
	if err != nil {
		return fmt.Errorf("query labs: %w", err)
	}
	printJSON(labs)

	fmt.Println()
	fmt.Println("--- Analytics: Risk Score P001 ---")
	riskTool := &analytics.RiskScoreTool{}
	v := vitals.(ehr.VitalSigns)
	l := labs.(ehr.LabResults)
	abnormalCount := 0
	for _, r := range l.Results {
		if r.Abnormal {
			abnormalCount++
		}
	}
	risk, err := riskTool.Execute(ctx, map[string]any{
		"patient_id":         "P001",
		"heart_rate":         float64(v.HeartRate),
		"systolic_bp":        float64(v.SystolicBP),
		"spo2":               float64(v.SpO2),
		"respiratory_rate":   float64(v.RespiratoryRate),
		"temperature":        v.Temperature,
		"abnormal_lab_count": float64(abnormalCount),
	})
	if err != nil {
		return fmt.Errorf("risk score: %w", err)
	}
	printJSON(risk)

	// --- Workflow layer demo ---
	fmt.Println()
	fmt.Println("--- Workflow: Patient Summary Plan ---")
	psWorkflow := &workflows.PatientSummaryWorkflow{}
	plan, err := workflows.BuildPlan(psWorkflow, map[string]any{"patient_id": "P001"})
	if err != nil {
		return fmt.Errorf("patient summary plan: %w", err)
	}
	printJSON(plan)

	fmt.Println()
	fmt.Println("--- Workflow: Shift Handover Plan ---")
	shWorkflow := &workflows.ShiftHandoverWorkflow{}
	shPlan, err := workflows.BuildPlan(shWorkflow, map[string]any{
		"unit":       "ICU",
		"patient_id": "P001",
	})
	if err != nil {
		return fmt.Errorf("shift handover plan: %w", err)
	}
	printJSON(shPlan)

	fmt.Println()
	fmt.Println("=== Demo complete ===")
	return nil
}

func printJSON(v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "json error: %v\n", err)
		return
	}
	fmt.Println(string(data))
}
