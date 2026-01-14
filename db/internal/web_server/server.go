package web_server

import (
	"encoding/json"
	"fmt"
	"minibank/internal/catalog"
	"minibank/internal/indexing"
	"minibank/internal/parser"
	"minibank/internal/planner"
	"minibank/internal/storage"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Server struct {
	Planner *planner.Planner
	Catalog *catalog.Catalog
	DataDir string
}

func NewServer(p *planner.Planner, c *catalog.Catalog, dataDir string) *Server {
	return &Server{Planner: p, Catalog: c, DataDir: dataDir}
}

func (s *Server) Start(port string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/users", s.handleUsers)
	mux.HandleFunc("/api/wallets", s.handleWallets)
	mux.HandleFunc("/api/wallets", s.handleWallets)
	mux.HandleFunc("/api/transactions", s.handleTransactions)
	mux.HandleFunc("/api/reports/user-wallets", s.handleReport)

	handler := s.enableCORS(mux)

	// Auto-initialize schema for the demo
	if err := s.ensureDemoSchema(); err != nil {
		fmt.Printf("Warning: failed to ensure demo schema: %v\n", err)
	}

	fmt.Printf("Web demo running on http://localhost%s\n", port)
	return http.ListenAndServe(port, handler)
}

func (s *Server) ensureDemoSchema() error {
	// Create users table
	if _, exists := s.Catalog.GetTable("users"); !exists {
		fmt.Println("Initializing 'users' table...")
		resp := s.executeQuery("CREATE TABLE users (id INT, name STRING, email STRING, PRIMARY KEY (id))")
		if resp.Error != "" {
			return fmt.Errorf("failed to create users table: %s", resp.Error)
		}
	}
	// Create wallets table
	if _, exists := s.Catalog.GetTable("wallets"); !exists {
		fmt.Println("Initializing 'wallets' table...")
		resp := s.executeQuery("CREATE TABLE wallets (id INT, user_id INT, balance DECIMAL, PRIMARY KEY (id))")
		if resp.Error != "" {
			return fmt.Errorf("failed to create wallets table: %s", resp.Error)
		}
	}
	// Create transactions table
	if _, exists := s.Catalog.GetTable("transactions"); !exists {
		fmt.Println("Initializing 'transactions' table...")
		resp := s.executeQuery("CREATE TABLE transactions (id INT, wallet_id INT, amount DECIMAL, type STRING, PRIMARY KEY (id))")
		if resp.Error != "" {
			return fmt.Errorf("failed to create transactions table: %s", resp.Error)
		}
	}
	return nil
}

func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	s.handleGenericCRUD(w, r, "users", "id", []string{"id", "name", "email"})
}

func (s *Server) handleWallets(w http.ResponseWriter, r *http.Request) {
	s.handleGenericCRUD(w, r, "wallets", "id", []string{"id", "user_id", "balance"})
}

func (s *Server) handleTransactions(w http.ResponseWriter, r *http.Request) {
	s.handleGenericCRUD(w, r, "transactions", "id", []string{"id", "wallet_id", "amount", "type"})
}

func (s *Server) handleReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// JOIN users + wallets
	sql := "SELECT users.id, users.name, wallets.balance FROM users JOIN wallets ON users.id = wallets.user_id"
	resp := s.executeQuery(sql)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleGenericCRUD(w http.ResponseWriter, r *http.Request, table string, pkCol string, columns []string) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		sql := fmt.Sprintf("SELECT * FROM %s", table)
		resp := s.executeQuery(sql)
		json.NewEncoder(w).Encode(resp)

	case "POST":
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var colsStr []string
		placeholders := make([]string, 0, len(columns))
		args := make([]interface{}, 0, len(columns))

		for _, col := range columns {
			if val, ok := body[col]; ok {
				colsStr = append(colsStr, col)
				placeholders = append(placeholders, "?")
				args = append(args, val)
			}
		}

		sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(colsStr, ", "), strings.Join(placeholders, ", "))
		resp := s.executePrepared(sql, args)
		json.NewEncoder(w).Encode(resp)

	case "PUT":
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		pkVal, ok := body[pkCol]
		if !ok {
			http.Error(w, "Missing primary key for update", http.StatusBadRequest)
			return
		}

		setStr := ""
		first := true
		for _, col := range columns {
			if col == pkCol {
				continue
			}
			if val, ok := body[col]; ok {
				if !first {
					setStr += ", "
				}

				switch v := val.(type) {
				case string:
					setStr += fmt.Sprintf("%s = '%s'", col, v)
				default:
					setStr += fmt.Sprintf("%s = %v", col, v)
				}
				first = false
			}
		}

		sql := fmt.Sprintf("UPDATE %s SET %s WHERE %s = %v", table, setStr, pkCol, pkVal)
		resp := s.executeQuery(sql)
		json.NewEncoder(w).Encode(resp)

	case "DELETE":
		pkValRaw := r.URL.Query().Get(pkCol)
		if pkValRaw == "" {
			http.Error(w, "Missing primary key for delete", http.StatusBadRequest)
			return
		}

		// Validate integer
		pkVal, err := strconv.Atoi(pkValRaw)
		if err != nil {
			http.Error(w, "Invalid primary key format", http.StatusBadRequest)
			return
		}

		sql := fmt.Sprintf("DELETE FROM %s WHERE %s = %d", table, pkCol, pkVal)
		resp := s.executeQuery(sql)
		json.NewEncoder(w).Encode(resp)

	case "OPTIONS":
		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := os.Getenv("CORS_ORIGIN")
		if origin == "" {
			origin = "http://localhost:3000"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type QueryRequest struct {
	Query string `json:"query"`
}

type QueryResponse struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	Error   string          `json:"error,omitempty"`
}

func (s *Server) executePrepared(sql string, args []interface{}) QueryResponse {
	// Simple interpolation helper
	// Replaces ? with values.
	// NOTE: This is a basic implementation for the demo.
	// Real SQL engines use the binder/parser for this.

	finalSQL := sql
	for _, arg := range args {
		var valStr string
		switch v := arg.(type) {
		case string:
			// Basic escaping
			safe := strings.ReplaceAll(v, "'", "''")
			valStr = fmt.Sprintf("'%s'", safe)
		default:
			valStr = fmt.Sprintf("%v", v)
		}
		finalSQL = strings.Replace(finalSQL, "?", valStr, 1)
	}
	return s.executeQuery(finalSQL)
}

func (s *Server) executeQuery(sql string) QueryResponse {
	lexer := parser.NewLexer(sql)
	p := parser.NewParser(lexer)
	ast, err := p.Parse()
	if err != nil {
		return QueryResponse{Error: err.Error()}
	}

	if createStmt, ok := ast.(*parser.CreateTableStmt); ok {
		err := s.Catalog.CreateTable(createStmt.TableName, createStmt.Columns)
		if err != nil {
			return QueryResponse{Error: err.Error()}
		}
		if err := s.Catalog.SaveToFile(filepath.Join(s.DataDir, "catalog.json")); err != nil {
			return QueryResponse{Error: fmt.Sprintf("failed to save catalog: %v", err)}
		}
		return QueryResponse{Columns: []string{"Result"}, Rows: [][]interface{}{{"Table Created"}}}
	}

	if createIdx, ok := ast.(*parser.CreateIndexStmt); ok {
		// handle Create Index similar to REPL
		table, exists := s.Catalog.GetTable(createIdx.TableName)
		if !exists {
			return QueryResponse{Error: fmt.Sprintf("table %s not found", createIdx.TableName)}
		}

		colIdx := -1
		for i, col := range table.Columns {
			if col.Name == createIdx.Column {
				colIdx = i
				break
			}
		}
		if colIdx == -1 {
			return QueryResponse{Error: fmt.Sprintf("column %s not found", createIdx.Column)}
		}

		idx := indexing.NewHashIndex()
		// Where do we get storage? Server struct doesn't have Storage directly, but Planner does.
		// Planner has Storage *storage.Engine
		hf, err := s.Planner.Storage.GetHeapFile(createIdx.TableName)
		if err != nil {
			return QueryResponse{Error: err.Error()}
		}

		iter := hf.Iterator()
		count := 0
		for {
			data, rid, err := iter.Next() // Update to use corrected signature returning rid struct?
			// Wait, Iterator.Next returns (bytes, RID, error). Correct.
			if err != nil {
				return QueryResponse{Error: err.Error()}
			}
			if data == nil {
				break
			}

			tuple, err := storage.DeserializeTuple(data, table.Columns)
			if err != nil {
				return QueryResponse{Error: err.Error()}
			}
			idx.Insert(tuple.Cells[colIdx].Value, rid)
			count++
		}

		key := fmt.Sprintf("%s.%s", createIdx.TableName, createIdx.Column)
		s.Planner.Indices[key] = idx

		// Persist
		// Persist
		err = s.Catalog.AddIndex(createIdx.TableName, catalog.IndexDef{
			Name:   createIdx.IndexName,
			Column: createIdx.Column,
			Type:   "HASH",
		})
		if err != nil {
			return QueryResponse{Error: err.Error()}
		}

		if err := s.Catalog.SaveToFile(filepath.Join(s.DataDir, "catalog.json")); err != nil {
			return QueryResponse{Error: fmt.Sprintf("failed to save catalog: %v", err)}
		}

		return QueryResponse{Columns: []string{"Result"}, Rows: [][]interface{}{{fmt.Sprintf("Index Created (%d entries)", count)}}}
	}

	iter, err := s.Planner.CreatePlan(ast)
	if err != nil {
		return QueryResponse{Error: err.Error()}
	}

	if err := iter.Open(); err != nil {
		return QueryResponse{Error: err.Error()}
	}
	defer iter.Close()

	schema := iter.Schema()
	cols := make([]string, len(schema))
	for i, c := range schema {
		cols[i] = c.Name
	}

	var rows [][]interface{}
	for {
		t, err := iter.Next()
		if err != nil {
			return QueryResponse{Error: err.Error()}
		}
		if t == nil {
			break
		}

		rowVals := make([]interface{}, len(t.Cells))
		for i, c := range t.Cells {
			rowVals[i] = c.Value
		}
		rows = append(rows, rowVals)
	}

	return QueryResponse{Columns: cols, Rows: rows}
}
