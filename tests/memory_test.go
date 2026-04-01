package tests

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/openbotstack/openbotstack-core/memory/abstraction"
)

// MockMemoryManager simulates an ephemeral multi-tier memory system.
type MockMemoryManager struct {
	shortTerm []abstraction.MemoryEntry
	longTerm  []abstraction.MemoryEntry
}

func (m *MockMemoryManager) StoreShortTerm(_ context.Context, entry abstraction.MemoryEntry) error {
	m.shortTerm = append(m.shortTerm, entry)
	return nil
}

func (m *MockMemoryManager) StoreLongTerm(_ context.Context, entry abstraction.MemoryEntry) error {
	m.longTerm = append(m.longTerm, entry)
	return nil
}

// RetrieveSimilar acts as a simple keyword search across both tiers for testing purposes
func (m *MockMemoryManager) RetrieveSimilar(_ context.Context, query string, limit int) ([]abstraction.MemoryEntry, error) {
	var results []abstraction.MemoryEntry

	// Search long term
	for _, entry := range m.longTerm {
		if strings.Contains(strings.ToLower(entry.Content), strings.ToLower(query)) {
			results = append(results, entry)
		}
	}
	
	// Search short term
	for _, entry := range m.shortTerm {
		if strings.Contains(strings.ToLower(entry.Content), strings.ToLower(query)) {
			results = append(results, entry)
		}
	}

	if limit > 0 && len(results) > limit {
		return results[:limit], nil
	}
	return results, nil
}

func (m *MockMemoryManager) RetrieveByTag(_ context.Context, tags []string, limit int) ([]abstraction.MemoryEntry, error) {
	return nil, nil // Not needed for current tests
}

func (m *MockMemoryManager) Forget(_ context.Context, id string) error {
	return nil // Not needed for current tests
}

func (m *MockMemoryManager) Summarize(_ context.Context, entries []abstraction.MemoryEntry) (abstraction.MemoryEntry, error) {
	return abstraction.MemoryEntry{}, nil // Not needed for current tests
}

func TestMemoryIntegration_StoreAndRetrieveShortTerm(t *testing.T) {
	ctx := context.Background()
	manager := &MockMemoryManager{}

	entry := abstraction.MemoryEntry{
		ID:        "msg-01",
		Content:   "Patient P001 requires increased oxygen flow.",
		CreatedAt: time.Now(),
		Tags:      []string{"clinical", "respiratory"},
	}

	err := manager.StoreShortTerm(ctx, entry)
	if err != nil {
		t.Fatalf("StoreShortTerm failed: %v", err)
	}

	results, err := manager.RetrieveSimilar(ctx, "oxygen", 5)
	if err != nil {
		t.Fatalf("RetrieveSimilar failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result for oxygen, got %d", len(results))
	}
	if results[0].ID != "msg-01" {
		t.Errorf("Expected msg-01, got %s", results[0].ID)
	}
}

func TestMemoryIntegration_StoreAndRetrieveLongTerm(t *testing.T) {
	ctx := context.Background()
	manager := &MockMemoryManager{}

	entry := abstraction.MemoryEntry{
		ID:        "doc-01",
		Content:   "Sepsis protocol requires broad-spectrum antibiotics within 1 hour.",
		CreatedAt: time.Now(),
		Tags:      []string{"protocol", "sepsis"},
	}

	err := manager.StoreLongTerm(ctx, entry)
	if err != nil {
		t.Fatalf("StoreLongTerm failed: %v", err)
	}

	results, err := manager.RetrieveSimilar(ctx, "antibiotics", 5)
	if err != nil {
		t.Fatalf("RetrieveSimilar failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result for antibiotics, got %d", len(results))
	}
	if results[0].ID != "doc-01" {
		t.Errorf("Expected doc-01, got %s", results[0].ID)
	}
}

func TestMemoryIntegration_VectorRetrievalLimits(t *testing.T) {
	ctx := context.Background()
	manager := &MockMemoryManager{}

	// Insert multiple entries that match "protocol"
	for i := 0; i < 5; i++ {
		_ = manager.StoreLongTerm(ctx, abstraction.MemoryEntry{
			ID:        "protocol-doc",
			Content:   "Standard protocol details here",
			CreatedAt: time.Now(),
		})
	}

	// Requesting limit 2 should return 2 entries max
	results, err := manager.RetrieveSimilar(ctx, "protocol", 2)
	if err != nil {
		t.Fatalf("RetrieveSimilar failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected exactly 2 results due to limit, got %d", len(results))
	}
}
