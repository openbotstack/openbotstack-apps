package ehr

import (
	"context"
	"fmt"

	cskills "github.com/openbotstack/openbotstack-core/control/skills"
)

// LabResult represents a single lab test result.
type LabResult struct {
	TestName  string  `json:"test_name"`
	Value     float64 `json:"value"`
	Unit      string  `json:"unit"`
	Reference string  `json:"reference"`
	Abnormal  bool    `json:"abnormal"`
}

// LabResults groups lab results for a patient.
type LabResults struct {
	PatientID string      `json:"patient_id"`
	Results   []LabResult `json:"results"`
	Timestamp string      `json:"timestamp"`
}

// stubLabs holds the in-memory lab results data.
var stubLabs = map[string]LabResults{
	"P001": {PatientID: "P001", Timestamp: "2026-03-15T06:00:00Z", Results: []LabResult{
		{TestName: "WBC", Value: 18.5, Unit: "10^9/L", Reference: "4.0-11.0", Abnormal: true},
		{TestName: "Lactate", Value: 4.2, Unit: "mmol/L", Reference: "0.5-2.0", Abnormal: true},
		{TestName: "Procalcitonin", Value: 8.5, Unit: "ng/mL", Reference: "<0.5", Abnormal: true},
		{TestName: "Creatinine", Value: 2.1, Unit: "mg/dL", Reference: "0.7-1.3", Abnormal: true},
	}},
	"P002": {PatientID: "P002", Timestamp: "2026-03-15T06:00:00Z", Results: []LabResult{
		{TestName: "WBC", Value: 7.2, Unit: "10^9/L", Reference: "4.0-11.0", Abnormal: false},
		{TestName: "Hemoglobin", Value: 11.5, Unit: "g/dL", Reference: "12.0-16.0", Abnormal: true},
		{TestName: "Creatinine", Value: 0.9, Unit: "mg/dL", Reference: "0.7-1.3", Abnormal: false},
	}},
	"P003": {PatientID: "P003", Timestamp: "2026-03-15T06:00:00Z", Results: []LabResult{
		{TestName: "Troponin-I", Value: 15.2, Unit: "ng/mL", Reference: "<0.04", Abnormal: true},
		{TestName: "CK-MB", Value: 85.0, Unit: "U/L", Reference: "0-25", Abnormal: true},
		{TestName: "BNP", Value: 1250.0, Unit: "pg/mL", Reference: "<100", Abnormal: true},
	}},
	"P004": {PatientID: "P004", Timestamp: "2026-03-15T06:00:00Z", Results: []LabResult{
		{TestName: "PaO2", Value: 55.0, Unit: "mmHg", Reference: "80-100", Abnormal: true},
		{TestName: "PaCO2", Value: 50.0, Unit: "mmHg", Reference: "35-45", Abnormal: true},
		{TestName: "pH", Value: 7.28, Unit: "", Reference: "7.35-7.45", Abnormal: true},
		{TestName: "WBC", Value: 12.8, Unit: "10^9/L", Reference: "4.0-11.0", Abnormal: true},
	}},
}

// QueryLabsTool queries lab results from the stub EHR.
type QueryLabsTool struct{}

// Name implements Tool.
func (t *QueryLabsTool) Name() string { return "ehr.query_labs" }

// Description implements Tool.
func (t *QueryLabsTool) Description() string {
	return "Query patient lab results from the Electronic Health Record system"
}

// InputSchema implements Tool.
func (t *QueryLabsTool) InputSchema() *cskills.JSONSchema {
	return &cskills.JSONSchema{
		Type: "object",
		Properties: map[string]*cskills.JSONSchema{
			"patient_id": {Type: "string"},
		},
		Required: []string{"patient_id"},
	}
}

// Execute implements Tool.
func (t *QueryLabsTool) Execute(_ context.Context, input map[string]any) (any, error) {
	patientID, ok := input["patient_id"].(string)
	if !ok || patientID == "" {
		return nil, fmt.Errorf("patient_id is required")
	}

	labs, exists := stubLabs[patientID]
	if !exists {
		return nil, fmt.Errorf("labs not found for patient: %s", patientID)
	}

	return labs, nil
}
