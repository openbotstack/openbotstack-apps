// Package tools defines the Tool interface and ToolRegistry for openbotstack-apps.
//
// A Tool is a low-level, stateless adapter that wraps an external capability
// (database query, API call, calculation). Tools are invoked by skills and
// workflows as part of an execution plan.
package tools

import (
	"context"
	"fmt"
	"sync"

	cskills "github.com/openbotstack/openbotstack-core/control/skills"
)

// Tool represents a low-level, stateless adapter for an external capability.
type Tool interface {
	// Name returns the unique identifier for this tool (e.g., "ehr.query_patient").
	Name() string

	// Description returns a human-readable explanation of what this tool does.
	Description() string

	// InputSchema returns the JSON Schema defining expected inputs.
	InputSchema() *cskills.JSONSchema

	// Execute runs the tool with the given input and returns structured output.
	Execute(ctx context.Context, input map[string]any) (any, error)
}

// Registry holds registered tools and provides lookup by name.
type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

// NewRegistry creates an empty tool registry.
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry.
// Returns an error if a tool with the same name is already registered.
func (r *Registry) Register(tool Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := tool.Name()
	if _, exists := r.tools[name]; exists {
		return fmt.Errorf("tool already registered: %s", name)
	}
	r.tools[name] = tool
	return nil
}

// Get retrieves a tool by name.
// Returns an error if the tool is not found.
func (r *Registry) Get(name string) (Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool not found: %s", name)
	}
	return tool, nil
}

// List returns all registered tool names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	return names
}
