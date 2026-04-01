package nursing

import (
	"fmt"
	"time"

	cskills "github.com/openbotstack/openbotstack-core/control/skills"
)

// SummarizeStatusSkill generates a clinical status summary for a patient
// by composing data from vitals and lab results.
type SummarizeStatusSkill struct{}

// ID implements Skill.
func (s *SummarizeStatusSkill) ID() string { return "nursing/summarize_status" }

// Name implements Skill.
func (s *SummarizeStatusSkill) Name() string { return "Summarize Patient Status" }

// Description implements Skill.
func (s *SummarizeStatusSkill) Description() string {
	return "Generate a clinical status summary for a patient based on current vital signs and lab results"
}

// InputSchema implements Skill.
func (s *SummarizeStatusSkill) InputSchema() *cskills.JSONSchema {
	return &cskills.JSONSchema{
		Type: "object",
		Properties: map[string]*cskills.JSONSchema{
			"patient_id": {Type: "string"},
		},
		Required: []string{"patient_id"},
	}
}

// OutputSchema implements Skill.
func (s *SummarizeStatusSkill) OutputSchema() *cskills.JSONSchema {
	return &cskills.JSONSchema{
		Type: "object",
		Properties: map[string]*cskills.JSONSchema{
			"patient_id": {Type: "string"},
			"summary":    {Type: "string"},
			"risk_level": {Type: "string"},
			"vitals":     {Type: "object"},
			"labs":       {Type: "object"},
		},
	}
}

// RequiredPermissions implements Skill.
func (s *SummarizeStatusSkill) RequiredPermissions() []string {
	return []string{"ehr:read", "analytics:read"}
}

// Timeout implements Skill.
func (s *SummarizeStatusSkill) Timeout() time.Duration { return defaultSkillTimeout }

// Validate implements Skill.
func (s *SummarizeStatusSkill) Validate() error {
	if s.ID() == "" {
		return fmt.Errorf("skill ID must not be empty")
	}
	return nil
}
