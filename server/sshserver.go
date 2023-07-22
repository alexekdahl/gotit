package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/google/uuid"
)

type SSHServer struct {
	tunnelStorer      TunnelStorer
	ssh               *ssh.Server
	authorizedKeysMap map[string]bool
	session           ssh.Session
}

func NewSSHServer(tunnelStorer TunnelStorer, port string) (*SSHServer, error) {
	if tunnelStorer == nil {
		return nil, fmt.Errorf("tunnelStorer cannot be nil")
	}

	if port == "" {
		return nil, fmt.Errorf("port cannot be empty")
	}

	s := &SSHServer{
		tunnelStorer: tunnelStorer,
		ssh: &ssh.Server{
			Addr: ":" + port,
		},
		authorizedKeysMap: make(map[string]bool),
	}

	s.ssh.PublicKeyHandler = s.keyHandler
	s.ssh.Handler = s.handleSSH

	return s, nil
}

func (s *SSHServer) loadAuthorizedKeys() error {
	absolutePath, err := filepath.Abs("authorized_keys")
	if err != nil {
		return fmt.Errorf("Failed to get absolute path, err: %v", err)
	}

	authorizedKeysBytes, err := os.ReadFile(absolutePath)
	if err != nil {
		return fmt.Errorf("Failed to load authorized_keys, err: %v", err)
	}

	for len(authorizedKeysBytes) > 0 {
		pubKey, _, _, rest, err := ssh.ParseAuthorizedKey(authorizedKeysBytes)
		if err != nil {
			return fmt.Errorf("Failed to parse authorized_keys, err: %v", err)
		}

		s.authorizedKeysMap[string(pubKey.Marshal())] = true
		authorizedKeysBytes = rest
	}

	if len(s.authorizedKeysMap) == 0 {
		return fmt.Errorf("no authorized keys loaded")
	}

	return nil
}

func (s *SSHServer) keyHandler(ctx ssh.Context, key ssh.PublicKey) bool {
	if len(s.authorizedKeysMap) != 0 {
		return s.authorizedKeysMap[string(key.Marshal())]
	}

	err := s.loadAuthorizedKeys()
	if err != nil {
		log.Printf("Error loading authorized keys: %v", err)
		return false
	}
	return s.authorizedKeysMap[string(key.Marshal())]
}

func getServerAddr() string {
	addr := os.Getenv("GOTIT_ADDR")
	if addr != "" {
		return addr
	}

	return "http://localhost:8080"
}

func (s *SSHServer) handleSSH(session ssh.Session) {
	log.Printf("New SSH connection from %s\n", session.RemoteAddr())

	s.session = session
	tunnelChan := make(chan Tunnel)
	tunnelID := uuid.New().String()
	s.tunnelStorer.Put(tunnelID, tunnelChan)

	s.writeWelcomeMsg()
	s.writeUrl(tunnelID)

	go func() {
		<-session.Context().Done()
		log.Printf("SSH connection from %s closed\n", session.RemoteAddr())
		s.tunnelStorer.Delete(tunnelID)
	}()

	cmd := session.Command()
	if len(cmd) == 1 {
		s.handleWithMime(session, cmd[0], tunnelID)
	} else {
		s.handleWithoutMime(session, tunnelID)
	}
}

func (s *SSHServer) handleWithMime(session ssh.Session, mime string, tunnelID string) {
	tunnelChan, ok := s.tunnelStorer.Get(tunnelID)
	if !ok {
		s.writeError()
		session.Context().Done()
	}

	tunnel := <-tunnelChan
	startTime := time.Now()

	filename := "gotit_file." + mime

	tunnel.w.Header().Set("Content-Type", mime)
	tunnel.w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	b, err := io.Copy(tunnel.w, session)
	if err != nil {
		log.Printf("Error copying data: %v", err)
		return
	}

	elapsedTime := time.Since(startTime).Seconds()
	speed := float64(b*8) / (elapsedTime * 1000000) // speed in Mb/s

	s.writeTransferSpeed(speed)

	tunnel.donech <- struct{}{}
}

// Send a welcome message to the user.
func (s *SSHServer) writeWelcomeMsg() {
	_, err := io.WriteString(s.session, fmt.Sprintf("Welcome, %s!\n", s.session.User()))
	if err != nil {
		log.Printf("Error writing to session: %v", err)
		return
	}
}

func (s *SSHServer) writeTypeUsage() {
	usage := "ssh gotit.sh [MIMETYPE] < response.json"

	_, err := io.WriteString(s.session, usage)
	if err != nil {
		log.Printf("Error writing to session: %v", err)
		return
	}
}

func (s *SSHServer) writeTransferSpeed(speed float64) {
	_, err := io.WriteString(s.session, fmt.Sprintf("Transfer speed: %.2f Mb/s\n", speed))
	if err != nil {
		log.Printf("Error writing to session: %v", err)
		return
	}
}

func (s *SSHServer) writeError() {
	error := "something went wrong"

	_, err := io.WriteString(s.session, error)
	if err != nil {
		log.Printf("Error writing to session: %v", err)
		return
	}
}

func (s *SSHServer) writeUrl(tunnelID string) {
	baseURL := getServerAddr()
	fullURL := baseURL + "/?id=" + tunnelID
	_, err := io.WriteString(s.session, fullURL+"\n")
	if err != nil {
		log.Printf("Error writing to session: %v", err)
		return
	}
}

func (s *SSHServer) handleWithoutMime(session ssh.Session, tunnelID string) {
	tunnelChan, ok := s.tunnelStorer.Get(tunnelID)
	if !ok {
		s.writeError()
		session.Context().Done()
	}

	tunnel := <-tunnelChan
	startTime := time.Now()

	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		_, err := io.Copy(pw, session)
		if err != nil {
			log.Printf("Error copying data: %v", err)
		}
	}()
	buf := make([]byte, 512)
	n, err := io.ReadFull(pr, buf)
	if err != nil && err != io.ErrUnexpectedEOF {
		log.Printf("Error reading data: %v", err)
		return
	}
	// Determine the MIME type from the first 512 bytes of the data
	mimeer := http.DetectContentType(buf[:n])

	ext, err := mime.ExtensionsByType(mimeer)
	fmt.Println("ext", ext)
	if err != nil {
		log.Printf("Error determining file extension: %v", err)
		return
	}
	// If the MIME type has associated extensions, use the last one
	filename := "gotit_file"
	if len(ext) > 0 {
		filename += ext[len(ext)-1]
	}
	// Set the Content-Disposition header to indicate the filename of the downloadable file
	tunnel.w.Header().Set("Content-Type", mimeer)
	tunnel.w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	// Write the buffered data to the tunnel's writer
	_, err = tunnel.w.Write(buf[:n])
	if err != nil {
		log.Printf("Error writing data to tunnel: %v", err)
		return
	}
	// Continue copying the rest of the data to the tunnel's writer
	b, err := io.Copy(tunnel.w, pr)
	if err != nil {
		log.Printf("Error copying data: %v", err)
		return
	}

	elapsedTime := time.Since(startTime).Seconds()
	speed := float64(b*8) / (elapsedTime * 1000000) // speed in Mb/s

	s.writeTransferSpeed(speed)

	tunnel.donech <- struct{}{}
}

func (s *SSHServer) StartSSHServer(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		if err := s.ssh.Close(); err != nil {
			log.Printf("Error closing SSH server: %v", err)
		}
	}()

	if err := s.ssh.ListenAndServe(); err != nil {
		log.Printf("SSH server ListenAndServe: %v", err)
		return err
	}

	return nil
}
