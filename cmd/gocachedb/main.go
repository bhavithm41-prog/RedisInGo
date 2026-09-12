package main

import (
	"fmt"

	"github.com/bhavithm41-prog/gocachedb/internal/store"
)

func main() {
	fmt.Println("GoCacheDB starting...")

	s := store.New()

	s.Set("name", "Bhavith")

	value, exists := s.Get("name")
	if exists {
		fmt.Println("GET name ->", value)
	} else {
		fmt.Println("GET name -> (nil)")
	}

	fmt.Println("EXISTS name ->", s.Exists("name"))
	fmt.Println("KEYS ->", s.Keys())

	deleted := s.Del("name")
	fmt.Println("DEL name -> deleted:", deleted)

	fmt.Println("EXISTS name ->", s.Exists("name"))
}
