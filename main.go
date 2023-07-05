package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
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

func (ts *TunnelStore) get(id string) (chan Tunnel, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	tunnel, ok := ts.tunnels[id]
	return tunnel, ok
}

func (ts *TunnelStore) put(id string, tunnel chan Tunnel) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.tunnels[id] = tunnel
}

func (ts *TunnelStore) delete(id string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	delete(ts.tunnels, id)
}

type Server struct {
	tunnelStore *TunnelStore
	server      http.Server
}

func (s *Server) startHTTPServer(ctx context.Context, port string) error {
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
	log.Printf("New HTTP connection from %s\n", r.RemoteAddr)

	id := r.URL.Query().Get("id")
	tunnelChan, ok := s.tunnelStore.get(id)
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
	s.tunnelStore.delete(id)
}

func (s *Server) handleSSH(session ssh.Session) {
	log.Printf("New SSH connection from %s\n", session.RemoteAddr())

	tunnelChan := make(chan Tunnel)
	tunnelID := uuid.New().String()
	s.tunnelStore.put(tunnelID, tunnelChan)

	// Send a welcome message to the user.
	_, err := io.WriteString(session, fmt.Sprintf("Welcome, %s!\n", session.User()))
	if err != nil {
		log.Printf("Error writing to session: %v", err)
		return
	}

	baseURL := "http://" + getLocalIP() + ":8080" // Now using local IP address
	fullURL := baseURL + "/?id=" + tunnelID
	_, err = io.WriteString(session, fullURL+"\n")
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

func keyHandler(ctx ssh.Context, key ssh.PublicKey) bool {
	absolutePath, err := filepath.Abs("keys") // dummy data
	if err != nil {
		log.Fatalf("Failed to get absolute path, err: %v", err)
	}

	authorizedKeysBytes, err := os.ReadFile(absolutePath)
	if err != nil {
		log.Fatalf("Failed to load authorized_keys, err: %v", err)
	}

	authorizedKeysMap := make(map[string]bool)
	for len(authorizedKeysBytes) > 0 {
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)
		if err != nil {
			log.Fatalf("Failed to parse authorized_keys, err: %v", err)
		}

		authorizedKeysMap[string(pubKey.Marshal())] = true
		authorizedKeysBytes = rest
	}

	return authorizedKeysMap[string(key.Marshal())]
}

func (s *Server) startSSHServer(ctx context.Context, port string) error {
	server := &ssh.Server{
		Addr:             ":" + port,
		Handler:          s.handleSSH,
		PublicKeyHandler: keyHandler,
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

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatalf("Cannot get local IP address: %v", err)
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return ""
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
		if err := server.startHTTPServer(ctx, "8080"); err != nil {
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
