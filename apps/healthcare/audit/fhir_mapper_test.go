package audit

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	coreaudit "github.com/openbotstack/openbotstack-core/audit"
)

func makeEnvelope() coreaudit.AuditEnvelope {
	return coreaudit.AuditEnvelope{
		EventID:     "evt-001",
		Timestamp:   time.Date(2025, 3, 15, 10, 30, 0, 0, time.UTC),
		TenantID:    "tenant-1",
		UserID:      "user-42",
		ExecutionID: "exec-100",
		StepID:      "step-5",
		EventType:   coreaudit.EventStepCompleted,
		Severity:    coreaudit.SeverityInfo,
		Source:      coreaudit.SourceExecutor,
		Action:      "skills.execute",
		Outcome:     "success",
		DurationMs:  150,
		Metadata:    map[string]string{"key": "value"},
	}
}

// --- NORMAL TESTS ---

func TestFHIRMapper_SingleEvent(t *testing.T) {
	m := FHIRAuditEventMapper{}
	env := makeEnvelope()

	result, err := m.Map(env)
	if err != nil {
		t.Fatalf("Map() error: %v", err)
	}

	event, ok := result.(FHIRAuditEvent)
	if !ok {
		t.Fatalf("Map() returned %T, want FHIRAuditEvent", result)
	}

	if event.ResourceType != "AuditEvent" {
		t.Errorf("ResourceType = %q, want %q", event.ResourceType, "AuditEvent")
	}
	if event.ID != "evt-001" {
		t.Errorf("ID = %q, want %q", event.ID, "evt-001")
	}
	if event.Action != "E" {
		t.Errorf("Action = %q, want %q", event.Action, "E")
	}
	if event.Outcome != "0" {
		t.Errorf("Outcome = %q, want %q (success)", event.Outcome, "0")
	}
	if event.Recorded != "2025-03-15T10:30:00Z" {
		t.Errorf("Recorded = %q, want ISO 8601", event.Recorded)
	}
	if event.Source.Site != "openbotstack" {
		t.Errorf("Source.Site = %q, want %q", event.Source.Site, "openbotstack")
	}
	if len(event.Agent) == 0 {
		t.Fatal("Agent is empty")
	}
	if event.Agent[0].Who.Reference != "Patient/user-42" {
		t.Errorf("Agent[0].Who.Reference = %q, want %q", event.Agent[0].Who.Reference, "Patient/user-42")
	}
	if len(event.Entity) == 0 {
		t.Fatal("Entity is empty")
	}
	if event.Entity[0].What.Reference != "Task/exec-100" {
		t.Errorf("Entity[0].What.Reference = %q, want %q", event.Entity[0].What.Reference, "Task/exec-100")
	}

	// Verify JSON serialization
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	if !strings.Contains(string(data), `"resourceType":"AuditEvent"`) {
		t.Errorf("JSON does not contain resourceType: %s", data)
	}
}

func TestFHIRMapper_BatchEvents(t *testing.T) {
	m := FHIRAuditEventMapper{}
	envs := []coreaudit.AuditEnvelope{
		makeEnvelope(),
		{
			EventID:   "evt-002",
			Timestamp: time.Date(2025, 3, 15, 11, 0, 0, 0, time.UTC),
			TenantID:  "tenant-1",
			UserID:    "user-99",
			EventType: coreaudit.EventPolicyDenied,
			Action:    "policy.enforce",
			Outcome:   "denied",
		},
	}

	result, err := m.MapBatch(envs)
	if err != nil {
		t.Fatalf("MapBatch() error: %v", err)
	}

	bundle, ok := result.(FHIRBundle)
	if !ok {
		t.Fatalf("MapBatch() returned %T, want FHIRBundle", result)
	}

	if bundle.ResourceType != "Bundle" {
		t.Errorf("ResourceType = %q, want %q", bundle.ResourceType, "Bundle")
	}
	if bundle.Type != "collection" {
		t.Errorf("Type = %q, want %q", bundle.Type, "collection")
	}
	if bundle.Total != 2 {
		t.Errorf("Total = %d, want 2", bundle.Total)
	}
	if len(bundle.Entry) != 2 {
		t.Fatalf("len(Entry) = %d, want 2", len(bundle.Entry))
	}
	if bundle.Entry[0].Resource.ID != "evt-001" {
		t.Errorf("Entry[0].ID = %q, want %q", bundle.Entry[0].Resource.ID, "evt-001")
	}
	if bundle.Entry[1].Resource.Outcome != "8" {
		t.Errorf("Entry[1].Outcome = %q, want %q (denied)", bundle.Entry[1].Resource.Outcome, "8")
	}
}

// --- ABNORMAL TESTS ---

func TestFHIRMapper_NilEnvelope(t *testing.T) {
	m := FHIRAuditEventMapper{}
	// Go does not allow nil for a value type, but we can test an empty struct
	var env coreaudit.AuditEnvelope
	_, err := m.Map(env)
	if err == nil {
		t.Error("Map() with zero-value envelope should return error")
	}
}

func TestFHIRMapper_EmptyEnvelope(t *testing.T) {
	m := FHIRAuditEventMapper{}
	env := coreaudit.AuditEnvelope{}
	_, err := m.Map(env)
	if err == nil {
		t.Error("Map() with empty envelope should return error")
	}
}

func TestFHIRMapper_EmptyBatch(t *testing.T) {
	m := FHIRAuditEventMapper{}
	result, err := m.MapBatch(nil)
	if err != nil {
		t.Fatalf("MapBatch(nil) error: %v", err)
	}
	bundle := result.(FHIRBundle)
	if bundle.Total != 0 {
		t.Errorf("Total = %d, want 0 for nil batch", bundle.Total)
	}
	if len(bundle.Entry) != 0 {
		t.Errorf("len(Entry) = %d, want 0", len(bundle.Entry))
	}

	// Also test empty slice
	result2, err2 := m.MapBatch([]coreaudit.AuditEnvelope{})
	if err2 != nil {
		t.Fatalf("MapBatch([]) error: %v", err2)
	}
	bundle2 := result2.(FHIRBundle)
	if bundle2.Total != 0 {
		t.Errorf("Total = %d, want 0 for empty batch", bundle2.Total)
	}
}

func TestFHIRMapper_BatchWithNilEnvelope(t *testing.T) {
	m := FHIRAuditEventMapper{}
	envs := []coreaudit.AuditEnvelope{
		makeEnvelope(),
		{}, // empty envelope should be skipped
		{EventID: "evt-003", Timestamp: time.Now()},
	}

	result, err := m.MapBatch(envs)
	if err != nil {
		t.Fatalf("MapBatch() error: %v", err)
	}
	bundle := result.(FHIRBundle)
	// Only 2 valid envelopes (makeEnvelope + evt-003)
	if bundle.Total != 2 {
		t.Errorf("Total = %d, want 2 (empty envelope skipped)", bundle.Total)
	}
}

func TestFHIRMapper_UnknownAction(t *testing.T) {
	m := FHIRAuditEventMapper{}
	env := makeEnvelope()
	env.EventType = "unknown.weird.event"
	env.Action = "something.strange"

	result, err := m.Map(env)
	if err != nil {
		t.Fatalf("Map() error: %v", err)
	}
	event := result.(FHIRAuditEvent)
	// Should fall back to "R" for unknown action
	if event.Action == "" {
		t.Error("Action should not be empty for unknown event type")
	}
}

func TestFHIRMapper_MissingUserID(t *testing.T) {
	m := FHIRAuditEventMapper{}
	env := makeEnvelope()
	env.UserID = ""

	result, err := m.Map(env)
	if err != nil {
		t.Fatalf("Map() error: %v", err)
	}
	event := result.(FHIRAuditEvent)
	if len(event.Agent) == 0 {
		t.Fatal("Agent is empty")
	}
	if event.Agent[0].Who.Reference != "Patient/unknown" {
		t.Errorf("Agent[0].Who.Reference = %q, want %q", event.Agent[0].Who.Reference, "Patient/unknown")
	}
}

func TestFHIRMapper_MissingTimestamp(t *testing.T) {
	m := FHIRAuditEventMapper{}
	env := makeEnvelope()
	env.Timestamp = time.Time{}

	result, err := m.Map(env)
	if err != nil {
		t.Fatalf("Map() error: %v", err)
	}
	event := result.(FHIRAuditEvent)
	// Should still produce valid recorded time (zero time formatted as ISO 8601)
	if event.Recorded == "" {
		t.Error("Recorded should not be empty")
	}
}

func TestFHIRMapper_VeryLargeBatch(t *testing.T) {
	m := FHIRAuditEventMapper{}
	const n = 1000
	envs := make([]coreaudit.AuditEnvelope, n)
	for i := range n {
		envs[i] = coreaudit.AuditEnvelope{
			EventID:   fmt.Sprintf("evt-%04d", i),
			Timestamp: time.Now(),
			TenantID:  "tenant-1",
			UserID:    "user-1",
			EventType: coreaudit.EventStepCompleted,
			Outcome:   "success",
		}
	}

	result, err := m.MapBatch(envs)
	if err != nil {
		t.Fatalf("MapBatch() error: %v", err)
	}
	bundle := result.(FHIRBundle)
	if bundle.Total != n {
		t.Errorf("Total = %d, want %d", bundle.Total, n)
	}
}

func TestFHIRMapper_SpecialCharactersInMetadata(t *testing.T) {
	m := FHIRAuditEventMapper{}
	env := makeEnvelope()
	env.Metadata = map[string]string{
		"unicode":    "éèê",
		"newline":    "line1\nline2",
		"quotes":     `"quoted"`,
		"backslash":  `C:\path\to\file`,
		"html":       `<script>alert("xss")</script>`,
	}

	result, err := m.Map(env)
	if err != nil {
		t.Fatalf("Map() error: %v", err)
	}
	// Verify JSON serialization handles special characters safely
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}
	if !json.Valid(data) {
		t.Error("result is not valid JSON")
	}
}

func TestFHIRMapper_FormatReturnsCorrectValue(t *testing.T) {
	m := FHIRAuditEventMapper{}
	if f := m.Format(); f != "fhir_auditevent" {
		t.Errorf("Format() = %q, want %q", f, "fhir_auditevent")
	}
}

func TestFHIRMapper_FailureOutcome(t *testing.T) {
	tests := []struct {
		name    string
		outcome string
		want    string
	}{
		{"success", "success", "0"},
		{"allowed", "allowed", "0"},
		{"failure", "failure", "8"},
		{"denied", "denied", "8"},
		{"timeout", "timeout", "8"},
		{"unknown", "weird", "4"},
		{"empty", "", "4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapOutcome(tt.outcome)
			if got != tt.want {
				t.Errorf("mapOutcome(%q) = %q, want %q", tt.outcome, got, tt.want)
			}
		})
	}
}

func TestFHIRMapper_ActionMapping(t *testing.T) {
	tests := []struct {
		name      string
		eventType coreaudit.EventType
		action    string
		want      string
	}{
		{"execution step", coreaudit.EventStepStarted, "skills.execute", "E"},
		{"execution completed", coreaudit.EventStepCompleted, "", "E"},
		{"execution failed", coreaudit.EventStepFailed, "", "E"},
		{"admin create", coreaudit.EventAdminProviderCreated, "", "C"},
		{"admin update", coreaudit.EventAdminProviderUpdated, "", "U"},
		{"admin delete", coreaudit.EventAdminKeyDeleted, "", "D"},
		{"admin enable", coreaudit.EventAdminSkillEnabled, "", "U"},
		{"admin disable", coreaudit.EventAdminSkillDisabled, "", "U"},
		{"admin reload", coreaudit.EventAdminSkillReloaded, "", "U"},
		{"policy allowed", coreaudit.EventPolicyAllowed, "", "E"},
		{"policy denied", coreaudit.EventPolicyDenied, "", "E"},
		{"system started", coreaudit.EventSystemStarted, "", "E"},
		{"action create", coreaudit.EventStepCompleted, "create", "C"},
		{"action read", coreaudit.EventStepCompleted, "read", "R"},
		{"action list", coreaudit.EventStepCompleted, "list_users", "R"},
		{"action get", coreaudit.EventStepCompleted, "get_data", "R"},
		{"action update", coreaudit.EventStepCompleted, "update", "U"},
		{"action delete", coreaudit.EventStepCompleted, "delete", "D"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapAction(tt.eventType, tt.action)
			if got != tt.want {
				t.Errorf("mapAction(%q, %q) = %q, want %q", tt.eventType, tt.action, got, tt.want)
			}
		})
	}
}

func TestFHIRMapper_ErrorField(t *testing.T) {
	m := FHIRAuditEventMapper{}
	env := makeEnvelope()
	env.Outcome = "failure"
	env.Error = "permission denied for resource"

	result, err := m.Map(env)
	if err != nil {
		t.Fatalf("Map() error: %v", err)
	}
	event := result.(FHIRAuditEvent)
	if event.Outcome != "8" {
		t.Errorf("Outcome = %q, want %q", event.Outcome, "8")
	}
	if event.OutcomeDesc != "permission denied for resource" {
		t.Errorf("OutcomeDesc = %q, want error message", event.OutcomeDesc)
	}
}

func TestFHIRMapper_MissingExecutionAndStepID(t *testing.T) {
	m := FHIRAuditEventMapper{}
	env := makeEnvelope()
	env.ExecutionID = ""
	env.StepID = ""

	result, err := m.Map(env)
	if err != nil {
		t.Fatalf("Map() error: %v", err)
	}
	event := result.(FHIRAuditEvent)
	if len(event.Entity) == 0 {
		t.Fatal("Entity should not be empty")
	}
	if event.Entity[0].What.Reference != "Task/unknown" {
		t.Errorf("Entity[0].What.Reference = %q, want %q", event.Entity[0].What.Reference, "Task/unknown")
	}
}

func TestFHIRMapper_StepIDOnly(t *testing.T) {
	m := FHIRAuditEventMapper{}
	env := makeEnvelope()
	env.ExecutionID = ""
	env.StepID = "step-only"

	result, err := m.Map(env)
	if err != nil {
		t.Fatalf("Map() error: %v", err)
	}
	event := result.(FHIRAuditEvent)
	if len(event.Entity) != 1 {
		t.Fatalf("len(Entity) = %d, want 1", len(event.Entity))
	}
	if event.Entity[0].What.Reference != "Task/step-only" {
		t.Errorf("Entity[0].What.Reference = %q, want %q", event.Entity[0].What.Reference, "Task/step-only")
	}
}

func TestFHIRMapper_InterfaceSatisfaction(t *testing.T) {
	// Verify FHIRAuditEventMapper satisfies the core interface.
	var _ coreaudit.AuditEventMapper = FHIRAuditEventMapper{}
}
