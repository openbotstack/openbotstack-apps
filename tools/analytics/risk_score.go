// Package analytics provides deterministic analytical tools.
package analytics

import (
	"context"
	"fmt"

	cskills "github.com/openbotstack/openbotstack-core/control/skills"
)

// Risk level constants.
const (
	RiskLow      = "low"
	RiskModerate = "moderate"
	RiskHigh     = "high"
	RiskCritical = "critical"
)

// Threshold constants for risk scoring.
const (
	heartRateHighThreshold       = 100
	systolicBPLowThreshold       = 90
	spo2LowThreshold             = 93
	respiratoryRateHighThreshold = 22
	temperatureHighThreshold     = 38.3

	abnormalLabsModerateThreshold = 2
	abnormalLabsHighThreshold     = 3
)

// RiskScoreResult contains the computed risk assessment.
type RiskScoreResult struct {
	PatientID    string  `json:"patient_id"`
	Score        float64 `json:"score"`
	Level        string  `json:"level"`
	Contributors []string `json:"contributors"`
}

// RiskScoreTool calculates a deterministic clinical risk score from vitals and labs.
type RiskScoreTool struct{}

// Name implements Tool.
func (t *RiskScoreTool) Name() string { return "analytics.risk_score" }

// Description implements Tool.
func (t *RiskScoreTool) Description() string {
	return "Calculate a deterministic risk score from patient vitals and lab results"
}

// InputSchema implements Tool.
func (t *RiskScoreTool) InputSchema() *cskills.JSONSchema {
	return &cskills.JSONSchema{
		Type: "object",
		Properties: map[string]*cskills.JSONSchema{
			"patient_id":      {Type: "string"},
			"heart_rate":      {Type: "number"},
			"systolic_bp":     {Type: "number"},
			"spo2":            {Type: "number"},
			"respiratory_rate": {Type: "number"},
			"temperature":     {Type: "number"},
			"abnormal_lab_count": {Type: "number"},
		},
		Required: []string{"patient_id"},
	}
}

// Execute implements Tool.
// Returns a RiskScoreResult with a deterministic score (0-100) and risk level.
func (t *RiskScoreTool) Execute(_ context.Context, input map[string]any) (any, error) {
	patientID, ok := input["patient_id"].(string)
	if !ok || patientID == "" {
		return nil, fmt.Errorf("patient_id is required")
	}

	var score float64
	var contributors []string

	// Heart rate contribution
	if hr, ok := toFloat64(input["heart_rate"]); ok && hr > float64(heartRateHighThreshold) {
		score += 15
		contributors = append(contributors, fmt.Sprintf("elevated HR (%.0f)", hr))
	}

	// Blood pressure contribution
	if sbp, ok := toFloat64(input["systolic_bp"]); ok && sbp < float64(systolicBPLowThreshold) {
		score += 20
		contributors = append(contributors, fmt.Sprintf("low SBP (%.0f)", sbp))
	}

	// SpO2 contribution
	if spo2, ok := toFloat64(input["spo2"]); ok && spo2 < float64(spo2LowThreshold) {
		score += 20
		contributors = append(contributors, fmt.Sprintf("low SpO2 (%.0f)", spo2))
	}

	// Respiratory rate contribution
	if rr, ok := toFloat64(input["respiratory_rate"]); ok && rr > float64(respiratoryRateHighThreshold) {
		score += 15
		contributors = append(contributors, fmt.Sprintf("elevated RR (%.0f)", rr))
	}

	// Temperature contribution
	if temp, ok := toFloat64(input["temperature"]); ok && temp > temperatureHighThreshold {
		score += 15
		contributors = append(contributors, fmt.Sprintf("fever (%.1f°C)", temp))
	}

	// Abnormal lab count contribution
	if labCount, ok := toFloat64(input["abnormal_lab_count"]); ok {
		if labCount >= float64(abnormalLabsHighThreshold) {
			score += 15
			contributors = append(contributors, fmt.Sprintf("%.0f abnormal labs", labCount))
		} else if labCount >= float64(abnormalLabsModerateThreshold) {
			score += 10
			contributors = append(contributors, fmt.Sprintf("%.0f abnormal labs", labCount))
		}
	}

	// Cap score at 100
	if score > 100 {
		score = 100
	}

	level := classifyRisk(score)

	return RiskScoreResult{
		PatientID:    patientID,
		Score:        score,
		Level:        level,
		Contributors: contributors,
	}, nil
}

// classifyRisk maps a numeric score to a risk level.
func classifyRisk(score float64) string {
	switch {
	case score >= 70:
		return RiskCritical
	case score >= 45:
		return RiskHigh
	case score >= 20:
		return RiskModerate
	default:
		return RiskLow
	}
}

// toFloat64 extracts a float64 from an any value (supports int, float32, float64).
func toFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	default:
		return 0, false
	}
}
