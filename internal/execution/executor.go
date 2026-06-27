package execution

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"vitess.io/vitess/go/vt/sqlparser"
)

// Executor handles safe execution of SQL queries
type Executor struct {
	db *sql.DB
}

// NewExecutor creates a new executor instance
func NewExecutor(dsn string) (*Executor, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return &Executor{db: db}, nil
}

// Execute runs a query after validating it is read-only
func (e *Executor) Execute(ctx context.Context, query string) ([]map[string]interface{}, error) {
	// 1. Validate the query
	if err := validateQuery(query); err != nil {
		return nil, fmt.Errorf("security validation failed: %w", err)
	}

	// 2. Enforce LIMIT if it's a SELECT statement
	finalQuery := enforceLimit(query)

	// 3. Execute with timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rows, err := e.db.QueryContext(ctx, finalQuery)
	if err != nil {
		return nil, fmt.Errorf("execution failed: %w", err)
	}
	defer rows.Close()

	// 4. Format results
	return scanRows(rows)
}

// validateQuery parses the SQL and ensures it is read-only
func validateQuery(query string) error {
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		return fmt.Errorf("invalid SQL syntax: %w", err)
	}

	switch stmt.(type) {
	case *sqlparser.Select, *sqlparser.Show, *sqlparser.Union, *sqlparser.Explain:
		return nil
	default:
		return fmt.Errorf("operation not allowed: only SELECT, SHOW, EXPLAIN are permitted")
	}
}

// enforceLimit appends LIMIT 1000 if it's a SELECT query and no limit exists
func enforceLimit(query string) string {
	trimmed := strings.TrimSpace(query)
	upper := strings.ToUpper(trimmed)

	// Simple check: if it's a SELECT and doesn't have a LIMIT clause
	if strings.HasPrefix(upper, "SELECT") && !strings.Contains(upper, "LIMIT") {
		return trimmed + " LIMIT 1000"
	}
	return trimmed
}

// scanRows converts sql.Rows to a slice of maps
func scanRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// Handle byte slices (common in MySQL driver)
			if b, ok := val.([]byte); ok {
				rowMap[col] = string(b)
			} else {
				rowMap[col] = val
			}
		}
		result = append(result, rowMap)
	}
	return result, nil
}
