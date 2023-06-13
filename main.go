package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"sync"

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

func startHTTPServer(port string) error {
	http.HandleFunc("/", handleReq)
	return http.ListenAndServe(":"+port, nil)
}

func startSSHServer(port string) error {
	ssh.Handle(handleSSH)
	return ssh.ListenAndServe(":"+port, nil)
}

func main() {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "3000" // Default port
	}

	sshPort := os.Getenv("SSH_PORT")
	if sshPort == "" {
		sshPort = "2222" // Default port
	}

	go func() {
		if err := startHTTPServer(httpPort); err != nil {
			log.Printf("Failed to start HTTP server: %v", err)
		}
	}()

	if err := startSSHServer(sshPort); err != nil {
		log.Printf("Failed to start SSH server: %v", err)
	}
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
