package ehr

import (
	"context"
	"fmt"

	cskills "github.com/openbotstack/openbotstack-core/control/skills"
)

// VitalSigns represents a set of vital sign measurements.
type VitalSigns struct {
	PatientID       string  `json:"patient_id"`
	HeartRate       int     `json:"heart_rate"`
	SystolicBP      int     `json:"systolic_bp"`
	DiastolicBP     int     `json:"diastolic_bp"`
	Temperature     float64 `json:"temperature"`
	RespiratoryRate int     `json:"respiratory_rate"`
	SpO2            int     `json:"spo2"`
	Timestamp       string  `json:"timestamp"`
}

// stubVitals holds the in-memory vital signs data.
var stubVitals = map[string]VitalSigns{
	"P001": {PatientID: "P001", HeartRate: 110, SystolicBP: 85, DiastolicBP: 55, Temperature: 38.9, RespiratoryRate: 24, SpO2: 92, Timestamp: "2026-03-15T08:00:00Z"},
	"P002": {PatientID: "P002", HeartRate: 78, SystolicBP: 120, DiastolicBP: 80, Temperature: 36.8, RespiratoryRate: 16, SpO2: 98, Timestamp: "2026-03-15T08:00:00Z"},
	"P003": {PatientID: "P003", HeartRate: 95, SystolicBP: 100, DiastolicBP: 65, Temperature: 37.2, RespiratoryRate: 20, SpO2: 94, Timestamp: "2026-03-15T08:00:00Z"},
	"P004": {PatientID: "P004", HeartRate: 105, SystolicBP: 90, DiastolicBP: 60, Temperature: 37.8, RespiratoryRate: 28, SpO2: 88, Timestamp: "2026-03-15T08:00:00Z"},
}

// QueryVitalsTool queries vital signs from the stub EHR.
type QueryVitalsTool struct{}

// Name implements Tool.
func (t *QueryVitalsTool) Name() string { return "ehr.query_vitals" }

// Description implements Tool.
func (t *QueryVitalsTool) Description() string {
	return "Query patient vital signs from the Electronic Health Record system"
}

// InputSchema implements Tool.
func (t *QueryVitalsTool) InputSchema() *cskills.JSONSchema {
	return &cskills.JSONSchema{
		Type: "object",
		Properties: map[string]*cskills.JSONSchema{
			"patient_id": {Type: "string"},
		},
		Required: []string{"patient_id"},
	}
}

// Execute implements Tool.
func (t *QueryVitalsTool) Execute(_ context.Context, input map[string]any) (any, error) {
	patientID, ok := input["patient_id"].(string)
	if !ok || patientID == "" {
		return nil, fmt.Errorf("patient_id is required")
	}

	vitals, exists := stubVitals[patientID]
	if !exists {
		return nil, fmt.Errorf("vitals not found for patient: %s", patientID)
	}

	return vitals, nil
}
