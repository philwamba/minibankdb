package main

import (
	"flag"
	"fmt"
	"minibank/internal/catalog"
	"minibank/internal/planner"
	"minibank/internal/repl"
	"minibank/internal/storage"
	"minibank/internal/web_server"
	"os"
)

func main() {
	mode := flag.String("mode", "repl", "Mode to run: 'repl' or 'server'")
	dataDir := flag.String("data", ".", "Data directory")
	port := flag.String("port", ":8080", "Server port")
	flag.Parse()

	cat := catalog.NewCatalog()
	if err := cat.LoadFromFile("catalog.json"); err != nil {
		// New catalog
	}

	store := storage.NewEngine(*dataDir)
	defer store.Close()

	switch *mode {
	case "repl":
		r := repl.NewREPL(cat, store)
		if err := r.Planner.RebuildIndices(); err != nil {
			fmt.Printf("Failed to rebuild indices: %v\n", err)
		}
		r.Run()
	case "server":
		pl := planner.NewPlanner(cat, store)
		if err := pl.RebuildIndices(); err != nil {
			fmt.Printf("Failed to rebuild indices: %v\n", err)
			os.Exit(1)
		}
		srv := web_server.NewServer(pl, cat)
		if err := srv.Start(*port); err != nil {
			fmt.Printf("Server failed: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown mode: %s\n", *mode)
		os.Exit(1)
	}
}
