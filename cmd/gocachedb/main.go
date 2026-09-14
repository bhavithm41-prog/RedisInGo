package main

import (
	"fmt"
	"os"

	"github.com/bhavithm41-prog/gocachedb/internal/metrics"
	"github.com/bhavithm41-prog/gocachedb/internal/server"
	"github.com/bhavithm41-prog/gocachedb/internal/store"
)

const (
	maxKeys  = 100
	dataFile = "dump.rdb"
)

func main() {
	m := metrics.New()
	s := store.New(maxKeys, m)

	if err := s.LoadFromFile(dataFile); err != nil {
		fmt.Println("Warning: failed to load", dataFile+":", err)
		fmt.Println("Starting with an empty database.")
	} else if _, statErr := os.Stat(dataFile); statErr == nil {
		fmt.Println("Data loaded from", dataFile)
	} else {
		fmt.Println("No existing data file found. Starting with an empty database.")
	}

	srv := server.New("6380", s, dataFile, m)

	if err := srv.Start(); err != nil {
		fmt.Println("Server error:", err)
		os.Exit(1)
	}
}
