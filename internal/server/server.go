package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/bhavithm41-prog/gocachedb/internal/command"
	"github.com/bhavithm41-prog/gocachedb/internal/store"
)

// Server holds everything our TCP server needs: the port to listen
// on, a command handler backed by the shared store, and a direct
// reference to that same store for background maintenance tasks
// like active expiration.
type Server struct {
	port    string
	handler *command.Handler
	store   *store.Store
}

// New creates a new Server bound to the given port, using the given store.
func New(port string, s *store.Store) *Server {
	return &Server{
		port:    port,
		handler: command.New(s),
		store:   s,
	}
}

// Start begins listening for TCP connections and blocks forever,
// accepting and handling clients as they connect. It also starts
// a background goroutine that actively expires timed-out keys.
func (srv *Server) Start() error {
	listener, err := net.Listen("tcp", ":"+srv.port)
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}
	defer listener.Close()

	fmt.Println("GoCacheDB server listening on port", srv.port)

	go srv.startExpirationCleanup(5 * time.Second)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go srv.handleConnection(conn)
	}
}

// handleConnection processes commands from a single client connection
// until the client disconnects.
func (srv *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Client connected:", conn.RemoteAddr())

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Client disconnected:", conn.RemoteAddr())
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		response := srv.handler.Execute(line)

		writer.WriteString(response + "\n")
		writer.Flush()
	}
}

// startExpirationCleanup runs store.CleanupExpired() on a fixed
// interval, forever, until the program exits. This is the "active
// expiration" half of TTL support — it complements the lazy
// expiration checks that already happen on every key access.
func (srv *Server) startExpirationCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		removed := srv.store.CleanupExpired()
		if removed > 0 {
			fmt.Println("Active expiration: removed", removed, "expired key(s)")
		}
	}
}
