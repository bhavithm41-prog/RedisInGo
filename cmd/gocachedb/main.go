package main

import (
	"fmt"
	"os"

	"github.com/bhavithm41-prog/gocachedb/internal/server"
	"github.com/bhavithm41-prog/gocachedb/internal/store"
)

func main() {
	s := store.New()

	srv := server.New("6380", s)

	if err := srv.Start(); err != nil {
		fmt.Println("Server error:", err)
		os.Exit(1)
	}
}
