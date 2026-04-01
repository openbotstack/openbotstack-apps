// Package workflows provides workflow definitions that compose skills and tools
// into multi-step execution plans.
//
// Workflows are the primary unit of orchestration in the application plane.
// Each workflow defines an ordered sequence of steps that reference skills
// or tools from the same repository.
//
// Workflows are descriptors — they produce ExecutionPlans that the runtime
// executes. Workflows themselves contain no execution logic.
package workflows

import (
	"fmt"
	"time"

	"github.com/openbotstack/openbotstack-core/execution"
)

// Workflow describes a multi-step process composed of skills and tools.
type Workflow interface {
	// ID returns a unique identifier for this workflow.
	ID() string

	// Name returns a human-readable name.
	Name() string

	// Description returns what this workflow does.
	Description() string

	// Steps returns the ordered sequence of execution steps.
	Steps(input map[string]any) ([]execution.ExecutionStep, error)

	// Timeout returns the maximum duration for the entire workflow.
	Timeout() time.Duration

	// Validate checks workflow configuration for correctness.
	Validate() error
}

// BuildPlan converts a Workflow into an ExecutionPlan ready for the runtime.
func BuildPlan(w Workflow, input map[string]any) (*execution.ExecutionPlan, error) {
	if err := w.Validate(); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	steps, err := w.Steps(input)
	if err != nil {
		return nil, fmt.Errorf("failed to build workflow steps: %w", err)
	}

	if len(steps) == 0 {
		return nil, fmt.Errorf("workflow %s produced zero steps", w.ID())
	}

	return &execution.ExecutionPlan{
		Steps: steps,
	}, nil
}
