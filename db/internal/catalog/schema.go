package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type ColumnType string

const (
	TypeInt       ColumnType = "INT"
	TypeString    ColumnType = "STRING"
	TypeDecimal   ColumnType = "DECIMAL"
	TypeBool      ColumnType = "BOOL"
	TypeTimestamp ColumnType = "TIMESTAMP"
)

type Column struct {
	Name      string     `json:"name"`
	Type      ColumnType `json:"type"`
	IsPrimary bool       `json:"is_primary"`
	IsUnique  bool       `json:"is_unique"`
	TableName string     `json:"-"`
}

type Table struct {
	Name    string   `json:"name"`
	Columns []Column `json:"columns"`
}

type Catalog struct {
	Tables map[string]*Table `json:"tables"`
	mu     sync.RWMutex
}

func NewCatalog() *Catalog {
	return &Catalog{
		Tables: make(map[string]*Table),
	}
}

func (c *Catalog) CreateTable(name string, columns []Column) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.Tables[name]; exists {
		return fmt.Errorf("table %s already exists", name)
	}

	c.Tables[name] = &Table{
		Name:    name,
		Columns: columns,
	}
	return nil
}

func (c *Catalog) GetTable(name string) (*Table, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	t, ok := c.Tables[name]
	return t, ok
}

func (c *Catalog) SaveToFile(path string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (c *Catalog) LoadFromFile(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return json.Unmarshal(data, c)
}
