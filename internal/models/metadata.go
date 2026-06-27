package models

// Column represents a single column in a database table
type Column struct {
	Name        string `json:"name"`
	DataType    string `json:"data_type"`
	Description string `json:"description"`
}

// Index represents a database index
type Index struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
}

// ForeignKey represents a relationship between tables
type ForeignKey struct {
	Name             string `json:"name"`
	Column           string `json:"column"`
	ReferencedTable  string `json:"referenced_table"`
	ReferencedColumn string `json:"referenced_column"`
}

// Table represents a database table and its associated columns, indexes, and relationships
type Table struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Columns     []Column     `json:"columns"`
	Indexes     []Index      `json:"indexes"`
	ForeignKeys []ForeignKey `json:"foreign_keys"`
}

// DatabaseSchema represents the full schema of a specific database
type DatabaseSchema struct {
	DatabaseName string  `json:"database_name"`
	Tables       []Table `json:"tables"`
}
