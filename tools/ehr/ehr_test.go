package ehr

import (
	"context"
	"testing"
)

func TestQueryPatientTool_ByID(t *testing.T) {
	tool := &QueryPatientTool{}

	result, err := tool.Execute(context.Background(), map[string]any{"patient_id": "P001"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	patient, ok := result.(PatientRecord)
	if !ok {
		t.Fatal("expected PatientRecord type")
	}
	if patient.Name != "Zhang Wei" {
		t.Errorf("expected Zhang Wei, got %s", patient.Name)
	}
}

func TestQueryPatientTool_ByUnit(t *testing.T) {
	tool := &QueryPatientTool{}

	result, err := tool.Execute(context.Background(), map[string]any{"unit": "ICU"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	patients, ok := result.([]PatientRecord)
	if !ok {
		t.Fatal("expected []PatientRecord type")
	}
	if len(patients) != 3 {
		t.Errorf("expected 3 ICU patients, got %d", len(patients))
	}
}

func TestQueryPatientTool_NotFound(t *testing.T) {
	tool := &QueryPatientTool{}

	_, err := tool.Execute(context.Background(), map[string]any{"patient_id": "INVALID"})
	if err == nil {
		t.Fatal("expected error for invalid patient_id")
	}
}

func TestQueryVitalsTool_Success(t *testing.T) {
	tool := &QueryVitalsTool{}

	result, err := tool.Execute(context.Background(), map[string]any{"patient_id": "P001"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	vitals, ok := result.(VitalSigns)
	if !ok {
		t.Fatal("expected VitalSigns type")
	}
	if vitals.HeartRate != 110 {
		t.Errorf("expected heart rate 110, got %d", vitals.HeartRate)
	}
}

func TestQueryVitalsTool_MissingID(t *testing.T) {
	tool := &QueryVitalsTool{}

	_, err := tool.Execute(context.Background(), map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing patient_id")
	}
}

func TestQueryLabsTool_Success(t *testing.T) {
	tool := &QueryLabsTool{}

	result, err := tool.Execute(context.Background(), map[string]any{"patient_id": "P001"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	labs, ok := result.(LabResults)
	if !ok {
		t.Fatal("expected LabResults type")
	}
	if len(labs.Results) != 4 {
		t.Errorf("expected 4 lab results, got %d", len(labs.Results))
	}
	// P001 has sepsis — all labs should be abnormal
	for _, r := range labs.Results {
		if !r.Abnormal {
			t.Errorf("expected abnormal result for %s", r.TestName)
		}
	}
}

func TestQueryLabsTool_NotFound(t *testing.T) {
	tool := &QueryLabsTool{}

	_, err := tool.Execute(context.Background(), map[string]any{"patient_id": "INVALID"})
	if err == nil {
		t.Fatal("expected error for invalid patient_id")
	}
}
