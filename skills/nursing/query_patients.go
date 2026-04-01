// Package nursing provides domain-specific nursing skills.
//
// Each skill implements the registry/skills.Skill interface from openbotstack-core.
// Skills are stateless descriptors — they provide metadata and schema definitions.
// Actual execution is handled by the runtime via the tools layer.
package nursing

import (
	"fmt"
	"time"

	cskills "github.com/openbotstack/openbotstack-core/control/skills"
)

const defaultSkillTimeout = 30 * time.Second

// QueryPatientsSkill retrieves the list of patients for a given unit.
type QueryPatientsSkill struct{}

// ID implements Skill.
func (s *QueryPatientsSkill) ID() string { return "nursing/query_patients" }

// Name implements Skill.
func (s *QueryPatientsSkill) Name() string { return "Query Patients" }

// Description implements Skill.
func (s *QueryPatientsSkill) Description() string {
	return "Retrieve the list of patients currently assigned to a nursing unit, including demographics and bed assignments"
}

// InputSchema implements Skill.
func (s *QueryPatientsSkill) InputSchema() *cskills.JSONSchema {
	return &cskills.JSONSchema{
		Type: "object",
		Properties: map[string]*cskills.JSONSchema{
			"unit": {Type: "string"},
		},
		Required: []string{"unit"},
	}
}

// OutputSchema implements Skill.
func (s *QueryPatientsSkill) OutputSchema() *cskills.JSONSchema {
	return &cskills.JSONSchema{
		Type: "object",
		Properties: map[string]*cskills.JSONSchema{
			"patients": {Type: "array"},
			"count":    {Type: "number"},
		},
	}
}

// RequiredPermissions implements Skill.
func (s *QueryPatientsSkill) RequiredPermissions() []string {
	return []string{"ehr:read"}
}

// Timeout implements Skill.
func (s *QueryPatientsSkill) Timeout() time.Duration { return defaultSkillTimeout }

// Validate implements Skill.
func (s *QueryPatientsSkill) Validate() error {
	if s.ID() == "" {
		return fmt.Errorf("skill ID must not be empty")
	}
	return nil
}
