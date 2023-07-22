package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/gliderlabs/ssh"
	"github.com/google/uuid"
)

type SSHServer struct {
	tunnelStorer TunnelStorer
	ssh          *ssh.Server
}

func NewSSHServer(tunnelStorer TunnelStorer, port string) *SSHServer {
	s := &SSHServer{
		tunnelStorer: tunnelStorer,
		ssh: &ssh.Server{
			Addr:             ":" + port,
			PublicKeyHandler: keyHandler,
		},
	}
	s.ssh.Handler = s.handleSSH

	return s
}

func keyHandler(ctx ssh.Context, key ssh.PublicKey) bool {
	absolutePath, err := filepath.Abs("authorized_keys") // dummy data
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

func getServerAddr() string {
	if os.Getenv("GOTIT_ADDR") != "" {
		return os.Getenv("GOTIT_ADDR")
	}

	return "http://localhost:8080"
}

func (s *SSHServer) handleSSH(session ssh.Session) {
	// Listen for the context done signal in a separate goroutine.

	log.Printf("New SSH connection from %s\n", session.RemoteAddr())

	tunnelChan := make(chan Tunnel)
	tunnelID := uuid.New().String()
	s.tunnelStorer.Put(tunnelID, tunnelChan)

	go func() {
		<-session.Context().Done()
		log.Printf("SSH connection from %s aborted\n", session.RemoteAddr())
		s.tunnelStorer.Delete(tunnelID)
	}()

	// Send a welcome message to the user.
	_, err := io.WriteString(session, fmt.Sprintf("Welcome, %s!\n", session.User()))
	if err != nil {
		log.Printf("Error writing to session: %v", err)
		return
	}

	baseURL := getServerAddr()
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

func (s *SSHServer) StartSSHServer(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		s.ssh.Close()
	}()

	if err := s.ssh.ListenAndServe(); err != nil {
		log.Printf("SSH server ListenAndServe: %v", err)
		return err
	}

	return nil
}
