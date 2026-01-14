package planner

import (
	"fmt"
	"minibank/internal/catalog"
	"minibank/internal/execution"
	"minibank/internal/indexing"
	"minibank/internal/parser"
	"minibank/internal/storage"
	"strconv"
)

type Planner struct {
	Catalog *catalog.Catalog
	Storage *storage.Engine
	Indices map[string]*indexing.HashIndex
}

func NewPlanner(cat *catalog.Catalog, store *storage.Engine) *Planner {
	return &Planner{Catalog: cat, Storage: store, Indices: make(map[string]*indexing.HashIndex)}
}

func (p *Planner) RebuildIndices() error {
	fmt.Println("Rebuilding indices...")
	for _, table := range p.Catalog.Tables {
		if len(table.Indexes) == 0 {
			continue
		}

		hf, err := p.Storage.GetHeapFile(table.Name)
		if err != nil {
			return err
		}

		// Create/Init indices in memory
		for _, idxDef := range table.Indexes {
			key := fmt.Sprintf("%s.%s", table.Name, idxDef.Column)
			p.Indices[key] = indexing.NewHashIndex()
			fmt.Printf("Initialized index: %s\n", key)
		}

		// Scan table and populate
		iter := hf.Iterator()
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

			for i, col := range table.Columns {
				// efficient lookup? loop is fine for rebuild
				key := fmt.Sprintf("%s.%s", table.Name, col.Name)
				if idx, ok := p.Indices[key]; ok {
					idx.Insert(tuple.Cells[i].Value, rid)
				}
			}
		}
	}
	return nil
}

func (p *Planner) CreatePlan(stmt parser.ASTNode) (execution.Iterator, error) {
	switch n := stmt.(type) {
	case *parser.SelectStmt:
		return p.planSelect(n)
	case *parser.InsertStmt:
		return p.planInsert(n)
	case *parser.UpdateStmt:
		return p.planUpdate(n)
	case *parser.DeleteStmt:
		return p.planDelete(n)
	}
	return nil, fmt.Errorf("unsupported statement type")
}

func (p *Planner) planSelect(stmt *parser.SelectStmt) (execution.Iterator, error) {
	table, exists := p.Catalog.GetTable(stmt.TableName)
	if !exists {
		return nil, fmt.Errorf("table %s not found", stmt.TableName)
	}

	hf, err := p.Storage.GetHeapFile(stmt.TableName)
	if err != nil {
		return nil, err
	}

	var root execution.Iterator

	usedIndex := false
	if stmt.Where != nil {
		if binExpr, ok := stmt.Where.Expr.(*parser.BinaryExpr); ok {
			if binExpr.Op == parser.OpEq {
				if ident, ok := binExpr.Left.(*parser.IdentifierExpr); ok {
					if lit, ok := binExpr.Right.(*parser.LiteralExpr); ok {
						var colType catalog.ColumnType
						for _, c := range table.Columns {
							if c.Name == ident.Name {
								colType = c.Type
								break
							}
						}

						val, err := castLiteral(lit.Value, colType)
						if err == nil {
							key := fmt.Sprintf("%s.%s", stmt.TableName, ident.Name)
							if idx, ok := p.Indices[key]; ok {
								fmt.Printf("[Planner] Using IndexScan on %s\n", key)
								root = execution.NewIndexScan(idx, hf, val, enrichSchema(table.Columns, stmt.TableName))
								usedIndex = true
							}
						}
					}
				}
			}
		}
	}

	if !usedIndex {
		root = execution.NewSeqScan(hf, enrichSchema(table.Columns, stmt.TableName))
	}

	if stmt.Join != nil {
		rightTable, exists := p.Catalog.GetTable(stmt.Join.Table)
		if !exists {
			return nil, fmt.Errorf("table %s not found", stmt.Join.Table)
		}
		hfRight, err := p.Storage.GetHeapFile(stmt.Join.Table)
		if err != nil {
			return nil, err
		}
		rightIter := execution.NewSeqScan(hfRight, enrichSchema(rightTable.Columns, stmt.Join.Table))

		joinOp := execution.NewNestedLoopJoin(root, rightIter, stmt.Join.On)
		root = joinOp
	}

	if stmt.Where != nil {
		root = execution.NewFilter(root, stmt.Where.Expr)
	}

	// Project
	if len(stmt.Fields) > 0 && stmt.Fields[0] != "*" {
		outSchema := []catalog.Column{}
		inSchema := root.Schema()
		for _, f := range stmt.Fields {
			found := false
			for _, col := range inSchema {
				if f == col.Name {
					outSchema = append(outSchema, col)
					found = true
					break
				}
				fqn := col.TableName + "." + col.Name
				if f == fqn {
					outSchema = append(outSchema, col)
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("column %s not found", f)
			}
		}
		root = execution.NewProject(root, stmt.Fields, outSchema)
	} else {
		// implicit *
	}

	return root, nil
}

func (p *Planner) planInsert(stmt *parser.InsertStmt) (execution.Iterator, error) {
	table, exists := p.Catalog.GetTable(stmt.TableName)
	if !exists {
		return nil, fmt.Errorf("table %s not found", stmt.TableName)
	}

	hf, err := p.Storage.GetHeapFile(stmt.TableName)
	if err != nil {
		return nil, err
	}

	rows := [][]interface{}{stmt.Values}

	return execution.NewInsert(hf, rows, table.Columns, p.Indices), nil
}

func (p *Planner) planUpdate(stmt *parser.UpdateStmt) (execution.Iterator, error) {
	table, exists := p.Catalog.GetTable(stmt.TableName)
	if !exists {
		return nil, fmt.Errorf("table %s not found", stmt.TableName)
	}
	hf, err := p.Storage.GetHeapFile(stmt.TableName)
	if err != nil {
		return nil, err
	}

	var root execution.Iterator = execution.NewSeqScan(hf, table.Columns)
	if stmt.Where != nil {
		root = execution.NewFilter(root, stmt.Where.Expr)
	}

	return execution.NewUpdate(hf, root, stmt.SetPairs), nil
}

func (p *Planner) planDelete(stmt *parser.DeleteStmt) (execution.Iterator, error) {
	table, exists := p.Catalog.GetTable(stmt.TableName)
	if !exists {
		return nil, fmt.Errorf("table %s not found", stmt.TableName)
	}
	hf, err := p.Storage.GetHeapFile(stmt.TableName)
	if err != nil {
		return nil, err
	}

	var root execution.Iterator = execution.NewSeqScan(hf, table.Columns)
	if stmt.Where != nil {
		root = execution.NewFilter(root, stmt.Where.Expr)
	}

	return execution.NewDelete(hf, root), nil
}

func enrichSchema(cols []catalog.Column, tableName string) []catalog.Column {
	newCols := make([]catalog.Column, len(cols))
	for i, c := range cols {
		newCols[i] = c
		newCols[i].TableName = tableName
	}
	return newCols
}

func castLiteral(val interface{}, targetType catalog.ColumnType) (interface{}, error) {
	if raw, ok := val.(parser.RawNumber); ok {
		sRaw := string(raw)
		switch targetType {
		case catalog.TypeInt:
			return strconv.ParseInt(sRaw, 10, 64)
		case catalog.TypeDecimal:
			return sRaw, nil
		default:
			return sRaw, nil
		}
	}
	return val, nil
}
