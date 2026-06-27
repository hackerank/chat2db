package models

// Column represents a single column in a database table
type Column struct {
	Name        string `json:"name"`
	DataType    string `json:"data_type"`
	Description string `json:"description"`
}

// Table represents a database table and its associated columns
type Table struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Columns     []Column `json:"columns"`
}

// DatabaseSchema represents the full schema of a specific database
type DatabaseSchema struct {
	DatabaseName string  `json:"database_name"`
	Tables       []Table `json:"tables"`
}
