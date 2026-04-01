package analytics

import (
	"context"
	"testing"
)

func TestRiskScoreTool_CriticalPatient(t *testing.T) {
	tool := &RiskScoreTool{}

	// P001-like: sepsis patient with multiple abnormalities
	result, err := tool.Execute(context.Background(), map[string]any{
		"patient_id":         "P001",
		"heart_rate":         float64(110),
		"systolic_bp":        float64(85),
		"spo2":               float64(92),
		"respiratory_rate":   float64(24),
		"temperature":        float64(38.9),
		"abnormal_lab_count": float64(4),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	riskResult, ok := result.(RiskScoreResult)
	if !ok {
		t.Fatal("expected RiskScoreResult type")
	}
	if riskResult.Level != RiskCritical {
		t.Errorf("expected critical risk, got %s (score: %.0f)", riskResult.Level, riskResult.Score)
	}
	if len(riskResult.Contributors) == 0 {
		t.Error("expected non-empty contributors")
	}
}

func TestRiskScoreTool_StablePatient(t *testing.T) {
	tool := &RiskScoreTool{}

	// P002-like: stable post-op patient
	result, err := tool.Execute(context.Background(), map[string]any{
		"patient_id":         "P002",
		"heart_rate":         float64(78),
		"systolic_bp":        float64(120),
		"spo2":               float64(98),
		"respiratory_rate":   float64(16),
		"temperature":        float64(36.8),
		"abnormal_lab_count": float64(1),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	riskResult, ok := result.(RiskScoreResult)
	if !ok {
		t.Fatal("expected RiskScoreResult type")
	}
	if riskResult.Level != RiskLow {
		t.Errorf("expected low risk, got %s (score: %.0f)", riskResult.Level, riskResult.Score)
	}
}

func TestRiskScoreTool_MissingPatientID(t *testing.T) {
	tool := &RiskScoreTool{}

	_, err := tool.Execute(context.Background(), map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing patient_id")
	}
}

func TestRiskScoreTool_ScoreCapped(t *testing.T) {
	tool := &RiskScoreTool{}

	// All indicators at worst possible values — score should cap at 100
	result, err := tool.Execute(context.Background(), map[string]any{
		"patient_id":         "PMAX",
		"heart_rate":         float64(200),
		"systolic_bp":        float64(40),
		"spo2":               float64(60),
		"respiratory_rate":   float64(40),
		"temperature":        float64(41.0),
		"abnormal_lab_count": float64(10),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	riskResult := result.(RiskScoreResult)
	if riskResult.Score > 100 {
		t.Errorf("score should not exceed 100, got %.0f", riskResult.Score)
	}
}

func TestRiskScoreTool_PartialInput(t *testing.T) {
	tool := &RiskScoreTool{}

	// Only heart rate provided — should still work
	result, err := tool.Execute(context.Background(), map[string]any{
		"patient_id": "PARTIAL",
		"heart_rate": float64(110),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	riskResult := result.(RiskScoreResult)
	if riskResult.Score != 15 {
		t.Errorf("expected score 15 for only elevated HR, got %.0f", riskResult.Score)
	}
	if riskResult.Level != RiskLow {
		t.Errorf("expected low risk for score 15, got %s", riskResult.Level)
	}
}

func TestClassifyRisk(t *testing.T) {
	tests := []struct {
		score    float64
		expected string
	}{
		{0, RiskLow},
		{19, RiskLow},
		{20, RiskModerate},
		{44, RiskModerate},
		{45, RiskHigh},
		{69, RiskHigh},
		{70, RiskCritical},
		{100, RiskCritical},
	}
	for _, tt := range tests {
		got := classifyRisk(tt.score)
		if got != tt.expected {
			t.Errorf("classifyRisk(%.0f): expected %s, got %s", tt.score, tt.expected, got)
		}
	}
}
