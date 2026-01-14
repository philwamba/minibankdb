package web_server

import (
	"encoding/json"
	"fmt"
	"minibank/internal/catalog"
	"minibank/internal/parser"
	"minibank/internal/planner"
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
	mux.HandleFunc("/api/query", s.handleQuery)
	mux.HandleFunc("/", s.handleStatic)

	handler := s.enableCORS(mux)

	fmt.Printf("Web demo running on http://localhost%s\n", port)
	return http.ListenAndServe(port, handler)
}

func (s *Server) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
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
