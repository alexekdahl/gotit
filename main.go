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
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/google/uuid"
)

type Tunnel struct {
	w      io.Writer
	donech chan struct{}
}

var (
	tunnels      = make(map[string]chan Tunnel)
	tunnelsMutex = &sync.RWMutex{}
)

func startHTTPServer(ctx context.Context, port string) error {
	srv := &http.Server{
		Addr: ":" + port,
	}

	http.HandleFunc("/", handleReq)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	<-ctx.Done()
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return srv.Shutdown(ctxShutDown)
}

func startSSHServer(ctx context.Context, port string) error {
	srv := &ssh.Server{
		Addr: ":" + port,
	}

	ssh.Handle(handleSSH)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if err.Error() != "ssh: Server closed" {
				log.Fatalf("ListenAndServe(): %v", err)
			}
		}
	}()

	<-ctx.Done()
	return srv.Close()
}

func main() {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "3000"
	}

	sshPort := os.Getenv("SSH_PORT")
	if sshPort == "" {
		sshPort = "2222"
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := startHTTPServer(ctx, httpPort); err != nil {
			log.Printf("Failed to start HTTP server: %v", err)
		}
	}()

	if err := startSSHServer(ctx, sshPort); err != nil {
		log.Printf("Failed to start SSH server: %v", err)
	}

	<-ctx.Done()

	log.Printf("Shutting down...")
}

func handleSSH(s ssh.Session) {
	id := uuid.New().String()
	ch := make(chan Tunnel)

	tunnelsMutex.Lock()
	tunnels[id] = ch
	tunnelsMutex.Unlock()

	log.Printf("Created tunnel with ID: %s", id)
	tunnel := <-ch
	log.Println("Tunnel is ready")

	_, err := io.Copy(tunnel.w, s)
	if err != nil {
		log.Printf("Failed to copy from session to tunnel: %v", err)
	}

	close(tunnel.donech)

	_, err = s.Write([]byte("done"))
	if err != nil {
		log.Printf("Failed to write to session: %v", err)
	}
}

func handleReq(w http.ResponseWriter, r *http.Request) {
	idstr := r.URL.Query().Get("id")
	if idstr == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	tunnelsMutex.RLock()
	tunnel, ok := tunnels[idstr]
	tunnelsMutex.RUnlock()
	if !ok {
		http.Error(w, "Tunnel not found", http.StatusNotFound)
		return
	}

	donech := make(chan struct{})
	tunnel <- Tunnel{
		w:      w,
		donech: donech,
	}

	<-donech
}
