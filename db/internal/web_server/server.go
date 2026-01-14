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
)

type Server struct {
	Planner *planner.Planner
	Catalog *catalog.Catalog
}

func NewServer(p *planner.Planner, c *catalog.Catalog) *Server {
	return &Server{Planner: p, Catalog: c}
}

func (s *Server) Start(port string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/users", s.handleUsers)
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

		colsStr := ""
		valsStr := ""
		first := true
		for _, col := range columns {
			if val, ok := body[col]; ok {
				if !first {
					colsStr += ", "
					valsStr += ", "
				}
				colsStr += col

				// Handle string quoting
				switch v := val.(type) {
				case string:
					valsStr += fmt.Sprintf("'%s'", v)
				default:
					valsStr += fmt.Sprintf("%v", v)
				}
				first = false
			}
		}

		sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, colsStr, valsStr)
		resp := s.executeQuery(sql)
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
		pkVal := r.URL.Query().Get(pkCol)
		if pkVal == "" {
			http.Error(w, "Missing primary key for delete", http.StatusBadRequest)
			return
		}

		sql := fmt.Sprintf("DELETE FROM %s WHERE %s = %s", table, pkCol, pkVal)
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
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
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

func (s *Server) handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := s.executeQuery(req.Query)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
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
		s.Catalog.SaveToFile("catalog.json")
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
		table.Indexes = append(table.Indexes, catalog.IndexDef{
			Name:   createIdx.IndexName,
			Column: createIdx.Column,
			Type:   "HASH",
		})
		s.Catalog.SaveToFile("catalog.json")

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

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}

	http.ServeFile(w, r, "../../../web/ui"+path)
}
