package chat

import (
	"chat2db/internal/execution"
	"chat2db/internal/generator"
	"chat2db/internal/models"
	"context"
	"fmt"
)

// Orchestrator coordinates the flow between the LLM generator and the SQL executor
type Orchestrator struct {
	generator *generator.Generator
	executor  *execution.Executor
}

// NewOrchestrator creates a new orchestrator instance
func NewOrchestrator(gen *generator.Generator, exec *execution.Executor) *Orchestrator {
	return &Orchestrator{
		generator: gen,
		executor:  exec,
	}
}

// ProcessPrompt handles the full pipeline: Generate SQL -> Validate/Execute -> Return Results
func (o *Orchestrator) ProcessPrompt(ctx context.Context, schema models.DatabaseSchema, userPrompt string) ([]map[string]interface{}, error) {
	// 1. Generate SQL using the LLM
	sqlQuery, err := o.generator.GenerateSQL(ctx, schema, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate SQL: %w", err)
	}

	// 2. Execute the generated SQL
	// The executor handles validation (read-only check) and safety (LIMIT/timeout)
	results, err := o.executor.Execute(ctx, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to execute generated SQL (%s): %w", sqlQuery, err)
	}

	return results, nil
}
