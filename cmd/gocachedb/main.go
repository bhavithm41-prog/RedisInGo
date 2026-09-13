package main

import (
	"fmt"
	"os"

	"github.com/bhavithm41-prog/gocachedb/internal/server"
	"github.com/bhavithm41-prog/gocachedb/internal/store"
)

const maxKeys = 100 // MAXMEMORY-equivalent: max number of keys before LRU eviction kicks in

func main() {
	s := store.New(maxKeys)

	srv := server.New("6380", s)

	if err := srv.Start(); err != nil {
		fmt.Println("Server error:", err)
		os.Exit(1)
	}
}
