package audit

import (
	"fmt"
	"strings"
	"time"

	coreaudit "github.com/openbotstack/openbotstack-core/audit"
)

// FHIR R4 types for AuditEvent mapping.

// FHIRCoding represents a coded value in FHIR.
type FHIRCoding struct {
	System string `json:"system,omitempty"`
	Code   string `json:"code,omitempty"`
	Display string `json:"display,omitempty"`
}

// FHIRPeriod represents a time period in FHIR.
type FHIRPeriod struct {
	Start string `json:"start,omitempty"`
	End   string `json:"end,omitempty"`
}

// FHIRReference represents a reference to another resource.
type FHIRReference struct {
	Reference string `json:"reference,omitempty"`
	Display   string `json:"display,omitempty"`
}

// FHIRAgent describes who performed the audited event.
type FHIRAgent struct {
	Who FHIRReference `json:"who"`
}

// FHIRSource describes the audit event reporter.
type FHIRSource struct {
	Site string       `json:"site,omitempty"`
	Observer FHIRReference `json:"observer"`
}

// FHIREntity describes an entity involved in the audited event.
type FHIREntity struct {
	What FHIRReference `json:"what"`
}

// FHIRAuditEvent represents a FHIR R4 AuditEvent resource.
type FHIRAuditEvent struct {
	ResourceType string       `json:"resourceType"` // always "AuditEvent"
	ID           string       `json:"id"`
	Type         FHIRCoding   `json:"type"`
	Subtype      []FHIRCoding `json:"subtype,omitempty"`
	Action       string       `json:"action"` // C/R/U/D/E
	Period       *FHIRPeriod  `json:"period,omitempty"`
	Recorded     string       `json:"recorded"` // ISO 8601
	Outcome      string       `json:"outcome"`  // 0=success, 4=minor, 8=serious, 12=major
	OutcomeDesc  string       `json:"outcomeDesc,omitempty"`
	Agent        []FHIRAgent  `json:"agent"`
	Source       FHIRSource   `json:"source"`
	Entity       []FHIREntity `json:"entity,omitempty"`
}

// FHIRBundle represents a FHIR Bundle resource containing AuditEvent entries.
type FHIRBundle struct {
	ResourceType string          `json:"resourceType"`
	Type         string          `json:"type"`
	Total        int             `json:"total"`
	Entry        []FHIRBundleEntry `json:"entry"`
}

// FHIRBundleEntry is a single entry in a FHIR Bundle.
type FHIRBundleEntry struct {
	Resource FHIRAuditEvent `json:"resource"`
}

// FHIRAuditEventMapper maps core AuditEnvelope to FHIR R4 AuditEvent resources.
type FHIRAuditEventMapper struct{}

// Format returns the unique format identifier for FHIR AuditEvent.
func (FHIRAuditEventMapper) Format() string {
	return "fhir_auditevent"
}

// Map converts a single audit envelope to a FHIR AuditEvent resource.
func (FHIRAuditEventMapper) Map(envelope coreaudit.AuditEnvelope) (any, error) {
	if envelope.EventID == "" && envelope.Timestamp.IsZero() {
		return nil, fmt.Errorf("audit envelope is empty")
	}

	event := FHIRAuditEvent{
		ResourceType: "AuditEvent",
		ID:           envelope.EventID,
		Type: FHIRCoding{
			System:  "http://terminology.hl7.org/CodeSystem/audit-event-type",
			Code:    "exec",
			Display: "Execution Event",
		},
		Subtype: mapSubtype(envelope.EventType),
		Action:  mapAction(envelope.EventType, envelope.Action),
		Recorded: envelope.Timestamp.UTC().Format(time.RFC3339Nano),
		Outcome:  mapOutcome(envelope.Outcome),
		Agent:    mapAgent(envelope.UserID),
		Source: FHIRSource{
			Site: "openbotstack",
			Observer: FHIRReference{
				Reference: "Device/openbotstack",
			},
		},
		Entity: mapEntity(envelope.ExecutionID, envelope.StepID),
	}

	if envelope.Error != "" {
		event.OutcomeDesc = envelope.Error
	}

	return event, nil
}

// MapBatch converts multiple envelopes to a FHIR Bundle of AuditEvent resources.
func (m FHIRAuditEventMapper) MapBatch(envelopes []coreaudit.AuditEnvelope) (any, error) {
	if len(envelopes) == 0 {
		return FHIRBundle{
			ResourceType: "Bundle",
			Type:         "collection",
			Total:        0,
			Entry:        []FHIRBundleEntry{},
		}, nil
	}

	entries := make([]FHIRBundleEntry, 0, len(envelopes))
	for i := range envelopes {
		if envelopes[i].EventID == "" && envelopes[i].Timestamp.IsZero() {
			continue
		}
		mapped, err := m.Map(envelopes[i])
		if err != nil {
			continue
		}
		entries = append(entries, FHIRBundleEntry{
			Resource: mapped.(FHIRAuditEvent),
		})
	}

	return FHIRBundle{
		ResourceType: "Bundle",
		Type:         "collection",
		Total:        len(entries),
		Entry:        entries,
	}, nil
}

// mapAction maps event types to FHIR action codes (C/R/U/D/E).
func mapAction(eventType coreaudit.EventType, action string) string {
	// Check explicit action first.
	switch {
	case strings.Contains(action, "create"), strings.Contains(action, "created"):
		return "C"
	case strings.Contains(action, "update"), strings.Contains(action, "updated"):
		return "U"
	case strings.Contains(action, "delete"), strings.Contains(action, "deleted"):
		return "D"
	case strings.Contains(action, "read"), strings.Contains(action, "list"), strings.Contains(action, "get"):
		return "R"
	}

	// Fall back to event type prefix.
	et := string(eventType)
	switch {
	case strings.HasPrefix(et, "execution."):
		return "E"
	case strings.HasPrefix(et, "admin."):
		if strings.Contains(et, "created") {
			return "C"
		}
		if strings.Contains(et, "updated") {
			return "U"
		}
		if strings.Contains(et, "deleted") {
			return "D"
		}
		if strings.Contains(et, "enabled") || strings.Contains(et, "disabled") || strings.Contains(et, "reloaded") {
			return "U"
		}
		return "U"
	case strings.HasPrefix(et, "policy."):
		return "E"
	case strings.HasPrefix(et, "system."):
		return "E"
	}
	return "R"
}

// mapOutcome maps envelope outcome to FHIR outcome code.
func mapOutcome(outcome string) string {
	switch outcome {
	case "success", "allowed":
		return "0"
	case "failure", "denied", "timeout":
		return "8"
	default:
		return "4" // minor failure / unknown
	}
}

// mapAgent creates FHIR agent entries from user info.
func mapAgent(userID string) []FHIRAgent {
	ref := "Patient/unknown"
	if userID != "" {
		ref = fmt.Sprintf("Patient/%s", userID)
	}
	return []FHIRAgent{
		{Who: FHIRReference{Reference: ref}},
	}
}

// mapEntity creates FHIR entity entries from execution/step info.
func mapEntity(executionID, stepID string) []FHIREntity {
	entities := make([]FHIREntity, 0, 2)
	if executionID != "" {
		entities = append(entities, FHIREntity{
			What: FHIRReference{Reference: fmt.Sprintf("Task/%s", executionID)},
		})
	}
	if stepID != "" {
		entities = append(entities, FHIREntity{
			What: FHIRReference{Reference: fmt.Sprintf("Task/%s", stepID)},
		})
	}
	if len(entities) == 0 {
		entities = append(entities, FHIREntity{
			What: FHIRReference{Reference: "Task/unknown"},
		})
	}
	return entities
}

// mapSubtype creates FHIR subtype codings from the event type.
func mapSubtype(eventType coreaudit.EventType) []FHIRCoding {
	et := string(eventType)
	if et == "" {
		return nil
	}
	return []FHIRCoding{
		{
			System:  "http://openbotstack.io/code-system/audit-event-subtype",
			Code:    et,
			Display: et,
		},
	}
}
