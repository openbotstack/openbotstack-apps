package workflows

import (
	"testing"
	"time"

	"github.com/openbotstack/openbotstack-core/execution"
)

func TestShiftHandoverWorkflow_BuildPlan(t *testing.T) {
	w := &ShiftHandoverWorkflow{}

	plan, err := BuildPlan(w, map[string]any{
		"unit":       "ICU",
		"patient_id": "P001",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Steps) != 5 {
		t.Errorf("expected 5 steps, got %d", len(plan.Steps))
	}
}

func TestShiftHandoverWorkflow_MissingUnit(t *testing.T) {
	w := &ShiftHandoverWorkflow{}

	_, err := BuildPlan(w, map[string]any{"patient_id": "P001"})
	if err == nil {
		t.Fatal("expected error for missing unit")
	}
}

func TestShiftHandoverWorkflow_MissingPatientID(t *testing.T) {
	w := &ShiftHandoverWorkflow{}

	_, err := BuildPlan(w, map[string]any{"unit": "ICU"})
	if err == nil {
		t.Fatal("expected error for missing patient_id")
	}
}

func TestPatientSummaryWorkflow_BuildPlan(t *testing.T) {
	w := &PatientSummaryWorkflow{}

	plan, err := BuildPlan(w, map[string]any{"patient_id": "P001"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plan.Steps) != 5 {
		t.Errorf("expected 5 steps, got %d", len(plan.Steps))
	}
}

func TestPatientSummaryWorkflow_MissingPatientID(t *testing.T) {
	w := &PatientSummaryWorkflow{}

	_, err := BuildPlan(w, map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing patient_id")
	}
}

func TestBuildPlan_ZeroSteps(t *testing.T) {
	w := &emptyWorkflow{}

	_, err := BuildPlan(w, map[string]any{})
	if err == nil {
		t.Fatal("expected error for zero steps")
	}
}

// --- test helpers ---

type emptyWorkflow struct{}

func (w *emptyWorkflow) ID() string          { return "test/empty" }
func (w *emptyWorkflow) Name() string        { return "Empty" }
func (w *emptyWorkflow) Description() string { return "test" }
func (w *emptyWorkflow) Validate() error     { return nil }

func (w *emptyWorkflow) Steps(_ map[string]any) ([]execution.ExecutionStep, error) {
	return nil, nil
}

func (w *emptyWorkflow) Timeout() time.Duration { return 0 }
