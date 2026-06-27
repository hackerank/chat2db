package schema

import (
	"chat2db/internal/models"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// DBConfig holds the connection details for a target database
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// SyncDatabase connects to a MySQL instance and extracts schema metadata
func SyncDatabase(cfg DBConfig) (*models.DatabaseSchema, error) {
	// Construct DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)
	
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	schema := &models.DatabaseSchema{
		DatabaseName: cfg.DBName,
		Tables:       []models.Table{},
	}

	// 1. Fetch all tables
	rows, err := db.Query("SELECT table_name, table_comment FROM information_schema.tables WHERE table_schema = ?", cfg.DBName)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var t models.Table
		if err := rows.Scan(&t.Name, &t.Description); err != nil {
			return nil, err
		}
		schema.Tables = append(schema.Tables, t)
	}

	// 2. Fetch columns for each table
	for i := range schema.Tables {
		cols, err := fetchColumns(db, cfg.DBName, schema.Tables[i].Name)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch columns for table %s: %w", schema.Tables[i].Name, err)
		}
		schema.Tables[i].Columns = cols
	}

	return schema, nil
}

// fetchColumns retrieves column metadata for a specific table
func fetchColumns(db *sql.DB, dbName, tableName string) ([]models.Column, error) {
	query := `
		SELECT column_name, data_type, column_comment 
		FROM information_schema.columns 
		WHERE table_schema = ? AND table_name = ?`
	
	rows, err := db.Query(query, dbName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []models.Column
	for rows.Next() {
		var c models.Column
		if err := rows.Scan(&c.Name, &c.DataType, &c.Description); err != nil {
			return nil, err
		}
		cols = append(cols, c)
	}
	return cols, nil
}
