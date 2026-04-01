package nursing

import (
	"fmt"
	"time"

	cskills "github.com/openbotstack/openbotstack-core/control/skills"
)

// GenerateSBARSkill generates an SBAR (Situation-Background-Assessment-Recommendation)
// handover communication for a patient.
//
// SBAR is a standardized framework used in healthcare communication:
//   - Situation: What is happening right now?
//   - Background: What is the clinical context?
//   - Assessment: What do I think the problem is?
//   - Recommendation: What do I suggest we do?
type GenerateSBARSkill struct{}

// ID implements Skill.
func (s *GenerateSBARSkill) ID() string { return "nursing/generate_sbar" }

// Name implements Skill.
func (s *GenerateSBARSkill) Name() string { return "Generate SBAR Handover" }

// Description implements Skill.
func (s *GenerateSBARSkill) Description() string {
	return "Generate a structured SBAR (Situation-Background-Assessment-Recommendation) handover communication for a patient"
}

// InputSchema implements Skill.
func (s *GenerateSBARSkill) InputSchema() *cskills.JSONSchema {
	return &cskills.JSONSchema{
		Type: "object",
		Properties: map[string]*cskills.JSONSchema{
			"patient_id": {Type: "string"},
		},
		Required: []string{"patient_id"},
	}
}

// OutputSchema implements Skill.
func (s *GenerateSBARSkill) OutputSchema() *cskills.JSONSchema {
	return &cskills.JSONSchema{
		Type: "object",
		Properties: map[string]*cskills.JSONSchema{
			"patient_id":     {Type: "string"},
			"situation":      {Type: "string"},
			"background":     {Type: "string"},
			"assessment":     {Type: "string"},
			"recommendation": {Type: "string"},
		},
	}
}

// RequiredPermissions implements Skill.
func (s *GenerateSBARSkill) RequiredPermissions() []string {
	return []string{"ehr:read", "analytics:read"}
}

// Timeout implements Skill.
func (s *GenerateSBARSkill) Timeout() time.Duration { return defaultSkillTimeout }

// Validate implements Skill.
func (s *GenerateSBARSkill) Validate() error {
	if s.ID() == "" {
		return fmt.Errorf("skill ID must not be empty")
	}
	return nil
}
