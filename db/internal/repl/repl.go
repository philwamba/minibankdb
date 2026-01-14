package repl

import (
	"bufio"
	"fmt"
	"minibank/internal/catalog"
	"minibank/internal/errors"
	"minibank/internal/execution"
	"minibank/internal/indexing"
	"minibank/internal/parser"
	"minibank/internal/planner"
	"minibank/internal/storage"
	"os"
	"strings"
)

type REPL struct {
	Catalog *catalog.Catalog
	Storage *storage.Engine
	Planner *planner.Planner
}

func NewREPL(cat *catalog.Catalog, store *storage.Engine) *REPL {
	return &REPL{
		Catalog: cat,
		Storage: store,
		Planner: planner.NewPlanner(cat, store),
	}
}

func (r *REPL) Run() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("MiniBankDB v0.1")
	fmt.Println("Type 'exit' to quit.")

	for {
		fmt.Print("minibank> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		input = strings.TrimSpace(input)

		if input == "exit" {
			break
		}
		if input == "" {
			continue
		}

		if err := r.Execute(input); err != nil {
			if dbErr, ok := err.(*errors.DBError); ok {
				fmt.Printf("Error: %s\n", dbErr.Message)
				if dbErr.Hint != "" {
					fmt.Printf("Hint: %s\n", dbErr.Hint)
				}
			} else {
				fmt.Printf("Error: %v\n", err)
			}
		}
	}
}

func (r *REPL) Execute(sql string) error {
	lexer := parser.NewLexer(sql)

	p := parser.NewParser(lexer)
	ast, err := p.Parse()
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	if createStmt, ok := ast.(*parser.CreateTableStmt); ok {
		return r.handleCreateTable(createStmt)
	}
	if createIdx, ok := ast.(*parser.CreateIndexStmt); ok {
		return r.handleCreateIndex(createIdx)
	}

	iter, err := r.Planner.CreatePlan(ast)
	if err != nil {
		return fmt.Errorf("plan error: %w", err)
	}

	return r.executeIter(iter)
}

func (r *REPL) handleCreateTable(stmt *parser.CreateTableStmt) error {
	err := r.Catalog.CreateTable(stmt.TableName, stmt.Columns)
	if err != nil {
		return err
	}
	_, err = r.Storage.GetHeapFile(stmt.TableName)
	if err != nil {
		return err
	}
	fmt.Println("CREATE TABLE")
	return r.Catalog.SaveToFile("catalog.json")
}

func (r *REPL) handleCreateIndex(stmt *parser.CreateIndexStmt) error {
	table, exists := r.Catalog.GetTable(stmt.TableName)
	if !exists {
		return fmt.Errorf("table %s not found", stmt.TableName)
	}

	colIdx := -1
	for i, col := range table.Columns {
		if col.Name == stmt.Column {
			colIdx = i
			break
		}
	}
	if colIdx == -1 {
		return fmt.Errorf("column %s not found in table %s", stmt.Column, stmt.TableName)
	}

	idx := indexing.NewHashIndex()
	hf, err := r.Storage.GetHeapFile(stmt.TableName)
	if err != nil {
		return err
	}

	fmt.Printf("Building index on %s.%s...\n", stmt.TableName, stmt.Column)
	iter := hf.Iterator()
	count := 0
	for {
		data, rid, err := iter.Next()
		if err != nil {
			return err
		}
		if data == nil {
			break
		}

		tuple, err := storage.DeserializeTuple(data, table.Columns)
		if err != nil {
			return err
		}

		val := tuple.Cells[colIdx].Value
		idx.Insert(val, rid)
		count++
	}

	key := fmt.Sprintf("%s.%s", stmt.TableName, stmt.Column)
	r.Planner.Indices[key] = idx

	// Persist
	table.Indexes = append(table.Indexes, catalog.IndexDef{
		Name:     stmt.IndexName,
		Column:   stmt.Column,
		Type:     "HASH",
		IsUnique: false,
	})
	r.Catalog.SaveToFile("catalog.json")

	fmt.Printf("CREATE INDEX (%d entries)\n", count)
	return nil
}

func (r *REPL) executeIter(it execution.Iterator) error {
	if err := it.Open(); err != nil {
		return err
	}
	defer it.Close()

	schema := it.Schema()
	for _, col := range schema {
		fmt.Printf("| %s\t", col.Name)
	}
	fmt.Println("|")
	fmt.Println(strings.Repeat("-", len(schema)*10))

	count := 0
	for {
		tuple, err := it.Next()
		if err != nil {
			return err
		}
		if tuple == nil {
			break
		}

		for _, cell := range tuple.Cells {
			fmt.Printf("| %v\t", cell.Value)
		}
		fmt.Println("|")
		count++
	}
	fmt.Printf("(%d rows)\n", count)
	return nil
}
