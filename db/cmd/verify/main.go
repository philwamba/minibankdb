package main

import (
	"fmt"
	"minibank/internal/catalog"
	"minibank/internal/repl"
	"minibank/internal/storage"
	"os"
)

func main() {
	os.Remove("catalog.json")
	files, _ := os.ReadDir(".")
	for _, f := range files {
		if len(f.Name()) > 5 && f.Name()[len(f.Name())-5:] == ".data" {
			os.Remove(f.Name())
		}
	}

	cat := catalog.NewCatalog()
	store := storage.NewEngine(".")
	r := repl.NewREPL(cat, store, ".")

	// Test Cases
	tests := []struct {
		name    string
		queries []string
		wantErr bool
	}{
		{
			name: "1. Decimal Literal Support",
			queries: []string{
				"CREATE TABLE t (id INT PRIMARY KEY, amt DECIMAL)",
				"INSERT INTO t VALUES (1, 10.50)",
				"SELECT * FROM t WHERE amt = 10.50",
			},
			wantErr: false,
		},
		{
			name: "2. Simple WHERE Boolean",
			queries: []string{
				"SELECT * FROM t WHERE id = 1",
			},
			wantErr: false,
		},
		{
			name: "3. Indexed WHERE",
			queries: []string{
				"CREATE INDEX idx_t_id ON t(id)",
				"SELECT * FROM t WHERE id = 1",
			},
			wantErr: false,
		},
		{
			name: "4. JOIN",
			queries: []string{
				"CREATE TABLE u (id INT PRIMARY KEY, email STRING UNIQUE)",
				"CREATE TABLE w (id INT PRIMARY KEY, user_id INT UNIQUE, balance DECIMAL)",
				"INSERT INTO u VALUES (1, 'a@b.com')",
				"INSERT INTO w VALUES (1, 1, 10.50)",
				"SELECT u.email, w.balance FROM u JOIN w ON u.id = w.user_id",
			},
			wantErr: false,
		},
		{
			name: "5. Decimal/Int Mixed",
			queries: []string{
				"INSERT INTO t VALUES (2, 20.50)",
			},
			wantErr: false,
		},
		{
			name: "6. Error Case: Int with Decimal",
			queries: []string{
				"INSERT INTO t VALUES (1.5, 10)",
			},
			wantErr: true,
		},
	}

	for _, t := range tests {
		fmt.Printf("Running Test: %s\n", t.name)
		for _, q := range t.queries {
			fmt.Printf("  Exec: %s\n", q)
			err := r.Execute(q)
			if err != nil {
				if t.wantErr {
					fmt.Printf("  Expected Error: %v\n", err)
				} else {
					fmt.Printf("  UNEXPECTED ERROR: %v\n", err)
					os.Exit(1)
				}
			} else if t.wantErr {
				fmt.Printf("  Expected Error but got Success\n")
				os.Exit(1)
			}
		}
		fmt.Println("  PASS")
	}
	fmt.Println("ALL TESTS PASSED")
}
