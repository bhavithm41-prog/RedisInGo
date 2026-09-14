package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/bhavithm41-prog/gocachedb/internal/command"
	"github.com/bhavithm41-prog/gocachedb/internal/metrics"
	"github.com/bhavithm41-prog/gocachedb/internal/store"
)

type Server struct {
	port    string
	handler *command.Handler
	store   *store.Store
	metrics *metrics.Metrics
}

func New(port string, s *store.Store, dataFile string, m *metrics.Metrics) *Server {
	return &Server{
		port:    port,
		handler: command.New(s, dataFile, m),
		store:   s,
		metrics: m,
	}
}

// Start binds to srv.port and serves forever. This is the normal
// entry point used by main.go.
func (srv *Server) Start() error {
	listener, err := net.Listen("tcp", ":"+srv.port)
	if err != nil {
		return fmt.Errorf("failed to start listener: %w", err)
	}
	return srv.Serve(listener)
}

// Serve accepts and handles connections on an already-created
// listener, and blocks forever. Splitting this out from Start lets
// tests supply their own listener — e.g. bound to port 0, so the OS
// assigns a free port and guarantees no collisions between tests.
func (srv *Server) Serve(listener net.Listener) error {
	defer listener.Close()

	fmt.Println("GoCacheDB server listening on", listener.Addr())

	go srv.startExpirationCleanup(5 * time.Second)

	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("accept error: %w", err)
		}
		go srv.handleConnection(conn)
	}
}

func (srv *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	srv.metrics.ClientConnected()
	defer srv.metrics.ClientDisconnected()

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

		writer.WriteString(response)
		writer.Flush()
	}
}

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
