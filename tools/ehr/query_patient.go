// Package ehr provides stub EHR (Electronic Health Record) tool adapters.
//
// All data is in-memory stub data for development and testing.
// Real implementations would connect to actual EHR systems.
package ehr

import (
	"context"
	"fmt"

	cskills "github.com/openbotstack/openbotstack-core/control/skills"
)

// PatientRecord represents a stub patient record.
type PatientRecord struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Age       int    `json:"age"`
	Gender    string `json:"gender"`
	Unit      string `json:"unit"`
	BedNumber string `json:"bed_number"`
	Diagnosis string `json:"diagnosis"`
}

// stubPatients holds the in-memory patient data.
var stubPatients = []PatientRecord{
	{ID: "P001", Name: "Zhang Wei", Age: 72, Gender: "M", Unit: "ICU", BedNumber: "ICU-01", Diagnosis: "Sepsis"},
	{ID: "P002", Name: "Li Na", Age: 45, Gender: "F", Unit: "ICU", BedNumber: "ICU-02", Diagnosis: "Post-operative monitoring"},
	{ID: "P003", Name: "Wang Jun", Age: 68, Gender: "M", Unit: "CCU", BedNumber: "CCU-05", Diagnosis: "Acute MI"},
	{ID: "P004", Name: "Chen Mei", Age: 55, Gender: "F", Unit: "ICU", BedNumber: "ICU-04", Diagnosis: "ARDS"},
}

// QueryPatientTool queries patient demographics from the stub EHR.
type QueryPatientTool struct{}

// Name implements Tool.
func (t *QueryPatientTool) Name() string { return "ehr.query_patient" }

// Description implements Tool.
func (t *QueryPatientTool) Description() string {
	return "Query patient demographics from the Electronic Health Record system"
}

// InputSchema implements Tool.
func (t *QueryPatientTool) InputSchema() *cskills.JSONSchema {
	return &cskills.JSONSchema{
		Type: "object",
		Properties: map[string]*cskills.JSONSchema{
			"patient_id": {Type: "string"},
			"unit":       {Type: "string"},
		},
	}
}

// Execute implements Tool.
// Accepts either "patient_id" (returns single) or "unit" (returns list).
func (t *QueryPatientTool) Execute(_ context.Context, input map[string]any) (any, error) {
	if patientID, ok := input["patient_id"].(string); ok && patientID != "" {
		for _, p := range stubPatients {
			if p.ID == patientID {
				return p, nil
			}
		}
		return nil, fmt.Errorf("patient not found: %s", patientID)
	}

	if unit, ok := input["unit"].(string); ok && unit != "" {
		var results []PatientRecord
		for _, p := range stubPatients {
			if p.Unit == unit {
				results = append(results, p)
			}
		}
		return results, nil
	}

	return stubPatients, nil
}
