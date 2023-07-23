package server

import (
	"context"
	"net/http"
	"time"

	"github.com/AlexEkdahl/gotit/utils/logger"
)

type HTTPServer struct {
	tunnelStorer TunnelStorer
	http         http.Server
	logger       logger.Logger
}

func NewHTTPServer(tunnelStorer TunnelStorer, logger logger.Logger, port string) *HTTPServer {
	s := &HTTPServer{
		tunnelStorer: tunnelStorer,
		http: http.Server{
			Addr: ":" + port,
		},
		logger: logger,
	}

	s.http.Handler = http.HandlerFunc(s.handleHTTPReq)

	return s
}

func (s *HTTPServer) StartHTTPServer(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		s.logger.Debug("Shuting down HTTP server")
		defer cancel()
		if err := s.http.Shutdown(shutdownCtx); err != nil {
			s.logger.Error(&HTTPTerminationError{
				err: err,
			})
		}
	}()

	if err := s.http.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *HTTPServer) handleHTTPReq(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	tunnelChan, ok := s.tunnelStorer.Get(id)
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		s.logger.Error(ErrCouldNotFindItem)
		return
	}

	tunnel := Tunnel{
		w:      w,
		donech: make(chan struct{}),
	}

	tunnelChan <- tunnel
	<-tunnel.donech
}
