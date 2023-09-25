package http

import (
	"context"
	"net/http"
	"time"

	"github.com/AlexEkdahl/gotit/pkg/pipe"
	"github.com/AlexEkdahl/gotit/pkg/util"
)

type TunnelStorer interface {
	Get(id string) (chan pipe.Tunnel, bool)
	Put(id string, tunnel chan pipe.Tunnel)
	Delete(id string)
}

type Server struct {
	tunnelStorer TunnelStorer
	http         http.Server
	logger       util.Logger
}

func NewServer(tunnelStorer TunnelStorer, logger util.Logger, port string) *Server {
	s := &Server{
		tunnelStorer: tunnelStorer,
		http: http.Server{
			Addr: ":" + port,
		},
		logger: logger,
	}

	s.http.Handler = http.HandlerFunc(s.handleReq)

	return s
}

func (s *Server) StartServer(ctx context.Context) error {
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

func (s *Server) handleReq(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	tunnelChan, ok := s.tunnelStorer.Get(id)
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		s.logger.Error(ErrCouldNotFindItem)
		return
	}

	tunnel := pipe.Tunnel{
		W:      w,
		Donech: make(chan struct{}),
	}

	tunnelChan <- tunnel
	<-tunnel.Donech
}
