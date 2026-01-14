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
	"path/filepath"
)

func main() {
	mode := flag.String("mode", "repl", "Mode to run: 'repl' or 'server'")
	dataDir := flag.String("data", ".", "Data directory")
	port := flag.String("port", ":8080", "Server port")
	flag.Parse()

	cat := catalog.NewCatalog()
	catalogPath := filepath.Join(*dataDir, "catalog.json")
	if err := cat.LoadFromFile(catalogPath); err != nil {
		if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Warning: failed to load catalog: %v (starting fresh)\n", err)
		}
	}

	store := storage.NewEngine(*dataDir)
	defer store.Close()

	switch *mode {
	case "repl":
		r := repl.NewREPL(cat, store, *dataDir)
		if err := r.Planner.RebuildIndices(); err != nil {
			fmt.Printf("Failed to rebuild indices: %v\n", err)
		}
		r.Run()
	case "server":
		pl := planner.NewPlanner(cat, store)
		if err := pl.RebuildIndices(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to rebuild indices: %v\n", err)
			os.Exit(1)
		}
		srv := web_server.NewServer(pl, cat, *dataDir)
		if err := srv.Start(*port); err != nil {
			fmt.Fprintf(os.Stderr, "Server failed: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown mode: %s\n", *mode)
		os.Exit(1)
	}
}
