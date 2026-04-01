package nursing

import (
	"testing"

	rskills "github.com/openbotstack/openbotstack-core/registry/skills"
)

// allSkills returns all nursing skills for table-driven tests.
func allSkills() []rskills.Skill {
	return []rskills.Skill{
		&QueryPatientsSkill{},
		&SummarizeStatusSkill{},
		&GenerateSBARSkill{},
	}
}

func TestSkills_ImplementInterface(t *testing.T) {
	for _, s := range allSkills() {
		// Verify the interface contract is satisfied
		if s.ID() == "" {
			t.Errorf("skill has empty ID")
		}
		if s.Name() == "" {
			t.Errorf("skill %s has empty Name", s.ID())
		}
		if s.Description() == "" {
			t.Errorf("skill %s has empty Description", s.ID())
		}
		if s.Timeout() <= 0 {
			t.Errorf("skill %s has non-positive Timeout", s.ID())
		}
	}
}

func TestSkills_HaveInputSchema(t *testing.T) {
	for _, s := range allSkills() {
		schema := s.InputSchema()
		if schema == nil {
			t.Errorf("skill %s has nil InputSchema", s.ID())
			continue
		}
		if schema.Type != "object" {
			t.Errorf("skill %s InputSchema type should be 'object', got '%s'", s.ID(), schema.Type)
		}
	}
}

func TestSkills_HaveOutputSchema(t *testing.T) {
	for _, s := range allSkills() {
		schema := s.OutputSchema()
		if schema == nil {
			t.Errorf("skill %s has nil OutputSchema", s.ID())
			continue
		}
		if schema.Type != "object" {
			t.Errorf("skill %s OutputSchema type should be 'object', got '%s'", s.ID(), schema.Type)
		}
	}
}

func TestSkills_Validate(t *testing.T) {
	for _, s := range allSkills() {
		if err := s.Validate(); err != nil {
			t.Errorf("skill %s Validate() failed: %v", s.ID(), err)
		}
	}
}

func TestSkills_RequirePermissions(t *testing.T) {
	for _, s := range allSkills() {
		perms := s.RequiredPermissions()
		if len(perms) == 0 {
			t.Errorf("skill %s has no required permissions", s.ID())
		}
	}
}

func TestQueryPatientsSkill_ID(t *testing.T) {
	s := &QueryPatientsSkill{}
	if s.ID() != "nursing/query_patients" {
		t.Errorf("unexpected ID: %s", s.ID())
	}
}

func TestSummarizeStatusSkill_ID(t *testing.T) {
	s := &SummarizeStatusSkill{}
	if s.ID() != "nursing/summarize_status" {
		t.Errorf("unexpected ID: %s", s.ID())
	}
}

func TestGenerateSBARSkill_ID(t *testing.T) {
	s := &GenerateSBARSkill{}
	if s.ID() != "nursing/generate_sbar" {
		t.Errorf("unexpected ID: %s", s.ID())
	}
}
