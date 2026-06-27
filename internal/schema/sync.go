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

	// 2. Fetch details for each table
	for i := range schema.Tables {
		tableName := schema.Tables[i].Name

		// Fetch Columns
		cols, err := fetchColumns(db, cfg.DBName, tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch columns for table %s: %w", tableName, err)
		}
		schema.Tables[i].Columns = cols

		// Fetch Indexes
		indexes, err := fetchIndexes(db, cfg.DBName, tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch indexes for table %s: %w", tableName, err)
		}
		schema.Tables[i].Indexes = indexes

		// Fetch Foreign Keys
		fks, err := fetchForeignKeys(db, cfg.DBName, tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch foreign keys for table %s: %w", tableName, err)
		}
		schema.Tables[i].ForeignKeys = fks
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

// fetchIndexes retrieves index metadata for a specific table
func fetchIndexes(db *sql.DB, dbName, tableName string) ([]models.Index, error) {
	query := `SHOW INDEX FROM ` + tableName + ` FROM ` + dbName
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexMap := make(map[string]*models.Index)
	for rows.Next() {
		var table, nonUnique, keyName, seqInIndex, columnName, collation, cardinality, subPart, packed, null, indexType, comment, indexComment, visible, expression string
		// MySQL SHOW INDEX returns 15 columns
		err := rows.Scan(&table, &nonUnique, &keyName, &seqInIndex, &columnName, &collation, &cardinality, &subPart, &packed, &null, &indexType, &comment, &indexComment, &visible, &expression)
		if err != nil {
			return nil, err
		}

		if _, ok := indexMap[keyName]; !ok {
			indexMap[keyName] = &models.Index{
				Name:   keyName,
				Unique: nonUnique == "0",
			}
		}
		indexMap[keyName].Columns = append(indexMap[keyName].Columns, columnName)
	}

	var indexes []models.Index
	for _, idx := range indexMap {
		indexes = append(indexes, *idx)
	}
	return indexes, nil
}

// fetchForeignKeys retrieves foreign key metadata for a specific table
func fetchForeignKeys(db *sql.DB, dbName, tableName string) ([]models.ForeignKey, error) {
	query := `
		SELECT CONSTRAINT_NAME, COLUMN_NAME, REFERENCED_TABLE_NAME, REFERENCED_COLUMN_NAME 
		FROM information_schema.KEY_COLUMN_USAGE 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND REFERENCED_TABLE_NAME IS NOT NULL`

	rows, err := db.Query(query, dbName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fks []models.ForeignKey
	for rows.Next() {
		var fk models.ForeignKey
		if err := rows.Scan(&fk.Name, &fk.Column, &fk.ReferencedTable, &fk.ReferencedColumn); err != nil {
			return nil, err
		}
		fks = append(fks, fk)
	}
	return fks, nil
}
