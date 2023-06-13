package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gliderlabs/ssh"
	"github.com/google/uuid"
)

type Tunnel struct {
	w      io.Writer
	donech chan struct{}
}

type TunnelStore struct {
	mu      sync.RWMutex
	tunnels map[string]chan Tunnel
}

func (ts *TunnelStore) Get(id string) (chan Tunnel, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	tunnel, ok := ts.tunnels[id]
	return tunnel, ok
}

func (ts *TunnelStore) Put(id string, tunnel chan Tunnel) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.tunnels[id] = tunnel
}

type Server struct {
	tunnelStore *TunnelStore
	server      http.Server
}

func (s *Server) StartHTTPServer(ctx context.Context, port string) error {
	s.server = http.Server{
		Addr:    ":" + port,
		Handler: http.HandlerFunc(s.handleReq),
	}

	go func() {
		<-ctx.Done()
		if err := s.server.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("HTTP server ListenAndServe: %v", err)
		return err
	}

	return nil
}

func (s *Server) handleReq(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	tunnelChan, ok := s.tunnelStore.Get(id)
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	tunnel := Tunnel{
		w:      w,
		donech: make(chan struct{}),
	}

	tunnelChan <- tunnel
	<-tunnel.donech
}

func (s *Server) handleSSH(session ssh.Session) {
	log.Printf("New SSH connection from %s\n", session.RemoteAddr())

	tunnelChan := make(chan Tunnel)
	tunnelID := uuid.New().String()

	s.tunnelStore.Put(tunnelID, tunnelChan)

	_, err := io.WriteString(session, tunnelID+"\n")
	if err != nil {
		log.Printf("Error writing to session: %v", err)
		return
	}

	tunnel := <-tunnelChan
	_, err = io.Copy(tunnel.w, session)
	if err != nil {
		log.Printf("Error copying data: %v", err)
		return
	}

	tunnel.donech <- struct{}{}
}

func (s *Server) startSSHServer(ctx context.Context, port string) error {
	server := &ssh.Server{
		Addr:    ":" + port,
		Handler: s.handleSSH,
	}

	go func() {
		<-ctx.Done()
		server.Close()
	}()

	if err := server.ListenAndServe(); err != nil {
		log.Printf("SSH server ListenAndServe: %v", err)
		return err
	}

	return nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigch
		cancel()
	}()

	var wg sync.WaitGroup
	wg.Add(2)

	tunnelStore := &TunnelStore{
		tunnels: make(map[string]chan Tunnel),
	}

	server := &Server{
		tunnelStore: tunnelStore,
	}

	go func() {
		defer wg.Done()
		if err := server.StartHTTPServer(ctx, "8080"); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := server.startSSHServer(ctx, "2222"); err != nil {
			log.Fatal(err)
		}
	}()

	wg.Wait()
}
