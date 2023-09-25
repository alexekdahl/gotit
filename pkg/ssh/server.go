package ssh

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/AlexEkdahl/gotit/pkg/pipe"
	"github.com/AlexEkdahl/gotit/pkg/util"
	"github.com/gliderlabs/ssh"
	"github.com/google/uuid"
)

type tunnelStorer interface {
	Get(id string) (chan pipe.Tunnel, bool)
	Put(id string, tunnel chan pipe.Tunnel)
	Delete(id string)
}

type SSHServer struct {
	tunnelStorer      tunnelStorer
	ssh               *ssh.Server
	authorizedKeysMap map[string]bool
	logger            util.Logger
	writer            *SessionWriter
	session           ssh.Session
	addr              string
}

const filename = "gotit"

func NewServer(ts tunnelStorer, logger util.Logger, port string) (*SSHServer, error) {
	if ts == nil {
		return nil, fmt.Errorf("tunnelStorer cannot be nil")
	}

	if port == "" {
		return nil, fmt.Errorf("port cannot be empty")
	}

	s := &SSHServer{
		tunnelStorer: ts,
		ssh: &ssh.Server{
			Addr: ":" + port,
		},
		authorizedKeysMap: make(map[string]bool),
		logger:            logger,
		addr:              getServerAddr(),
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
		s.logger.Warn("Error loading authorized keys: %v", err)
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
	startTime := time.Now()
	s.logger.Info("New SSH connection from %v@%v\n", session.User(), session.RemoteAddr())

	s.session = session
	s.writer = NewSessionWriter(session)

	tunnelChan := make(chan pipe.Tunnel)
	tunnelID := uuid.New().String()
	s.tunnelStorer.Put(tunnelID, tunnelChan)

	s.writer.WriteWelcomeMsg(session.User())
	s.writer.WriteURL(fmt.Sprintf("%s/?id=%s", s.addr, tunnelID))

	go func() {
		<-session.Context().Done()
		s.logger.Debug("SSH connection from %s closed\n", session.RemoteAddr())
		s.tunnelStorer.Delete(tunnelID)
	}()

	cmd := session.Command()
	if len(cmd) == 1 {
		err := s.handleWithMime(session, cmd[0], tunnelID)
		if err != nil {
			s.logger.Error(err)
			s.writer.WriteError(err)
			return
		}
	} else {
		err := s.handleWithoutMime(session, tunnelID)
		if err != nil {
			s.logger.Error(err)
			s.writer.WriteError(err)
			return
		}
	}
	elapsedTime := time.Since(startTime).Seconds()
	s.logger.Info("Transfer completed for %v@%v", session.User(), session.RemoteAddr())
	s.logger.Info("Connection opened %0.fs", elapsedTime)
}

func (s *SSHServer) handleWithMime(session ssh.Session, mime string, tunnelID string) error {
	tunnelChan, ok := s.tunnelStorer.Get(tunnelID)
	if !ok {
		s.writer.WriteError(ErrSomethingWentWrong)
		session.Context().Done()
	}

	tunnel := <-tunnelChan
	startTime := time.Now()

	fName := fmt.Sprintf("%s.%s", filename, mime)

	tunnel.W.Header().Set("Content-Type", mime)
	tunnel.W.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fName))

	b, err := io.Copy(tunnel.W, session)
	if err != nil {
		return &CopyDataError{err: err}
	}

	elapsedTime := time.Since(startTime).Seconds()
	speed := float64(b*8) / (elapsedTime * 1000000) // speed in Mb/s

	s.writer.WriteTransferDone(speed)

	tunnel.Donech <- struct{}{}

	return nil
}

func (s *SSHServer) handleWithoutMime(session ssh.Session, tunnelID string) error {
	tunnelChan, ok := s.tunnelStorer.Get(tunnelID)
	if !ok {
		s.writer.WriteError(ErrSomethingWentWrong)
		session.Context().Done()
	}

	tunnel := <-tunnelChan
	startTime := time.Now()

	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()
		_, err := io.Copy(pw, session)
		if err != nil {
			s.logger.Error(&CopyDataError{err: err})
		}
	}()
	buf := make([]byte, 512)
	n, err := io.ReadFull(pr, buf)
	if err != nil && err != io.ErrUnexpectedEOF {
		return &ReadDataError{err: err}
	}
	// Determine the MIME type from the first 512 bytes of the data
	mimeer := http.DetectContentType(buf[:n])

	ext, err := mime.ExtensionsByType(mimeer)
	if err != nil {
		return &FileExtensionError{err: err}
	}
	// If the MIME type has associated extensions, use the last one
	fName := filename
	if len(ext) > 0 {
		fName += ext[len(ext)-1]
	}
	// Set the Content-Disposition header to indicate the filename of the downloadable file
	tunnel.W.Header().Set("Content-Type", mimeer)
	tunnel.W.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fName))
	// Write the buffered data to the tunnel's writer
	_, err = tunnel.W.Write(buf[:n])
	if err != nil {
		return &TunnelWriteError{err: err}
	}
	// Continue copying the rest of the data to the tunnel's writer
	b, err := io.Copy(tunnel.W, pr)
	if err != nil {
		return &CopyDataError{err: err}
	}

	elapsedTime := time.Since(startTime).Seconds()
	speed := float64(b*8) / (elapsedTime * 1000000) // speed in Mb/s

	s.writer.WriteTransferDone(speed)

	tunnel.Donech <- struct{}{}

	return nil
}

func (s *SSHServer) StartServer(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		s.logger.Debug("Shuting down SSH server")
		defer cancel()
		if err := s.ssh.Shutdown(shutdownCtx); err != nil {
			s.logger.Error(&SSHTerminationError{err: err})
		}
	}()

	if err := s.ssh.ListenAndServe(); err != ssh.ErrServerClosed {
		return err
	}

	return nil
}
