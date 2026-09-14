package main

import (
	"fmt"
	"os"

	"github.com/bhavithm41-prog/gocachedb/internal/server"
	"github.com/bhavithm41-prog/gocachedb/internal/store"
)

const (
	maxKeys  = 100        // MAXMEMORY-equivalent: max number of keys before LRU eviction kicks in
	dataFile = "dump.rdb" // snapshot file used by SAVE/BGSAVE and loaded on startup
)

func main() {
	s := store.New(maxKeys)

	if err := s.LoadFromFile(dataFile); err != nil {
		fmt.Println("Warning: failed to load", dataFile+":", err)
		fmt.Println("Starting with an empty database.")
	} else if _, statErr := os.Stat(dataFile); statErr == nil {
		fmt.Println("Data loaded from", dataFile)
	} else {
		fmt.Println("No existing data file found. Starting with an empty database.")
	}

	srv := server.New("6380", s, dataFile)

	if err := srv.Start(); err != nil {
		fmt.Println("Server error:", err)
		os.Exit(1)
	}
}
