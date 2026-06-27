package chat

import (
	"chat2db/internal/execution"
	"chat2db/internal/formatter"
	"chat2db/internal/generator"
	"chat2db/internal/models"
	"context"
	"fmt"
)

// Orchestrator coordinates the flow between the LLM generator, SQL executor, and result formatter
type Orchestrator struct {
	generator *generator.Generator
	executor  *execution.Executor
	formatter *formatter.Formatter
}

// NewOrchestrator creates a new orchestrator instance
func NewOrchestrator(gen *generator.Generator, exec *execution.Executor, fmt *formatter.Formatter) *Orchestrator {
	return &Orchestrator{
		generator: gen,
		executor:  exec,
		formatter: fmt,
	}
}

// ProcessPrompt handles the full pipeline: Generate SQL -> Validate/Execute -> Format Results
func (o *Orchestrator) ProcessPrompt(ctx context.Context, schema models.DatabaseSchema, userPrompt string) (string, error) {
	// 1. Generate SQL using the LLM
	sqlQuery, err := o.generator.GenerateSQL(ctx, schema, userPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate SQL: %w", err)
	}

	// 2. Execute the generated SQL
	results, err := o.executor.Execute(ctx, sqlQuery)
	if err != nil {
		return "", fmt.Errorf("failed to execute generated SQL (%s): %w", sqlQuery, err)
	}

	// 3. Format the results into a natural language response
	summary, err := o.formatter.FormatResult(ctx, userPrompt, results)
	if err != nil {
		return "", fmt.Errorf("failed to format results: %w", err)
	}

	return summary, nil
}
