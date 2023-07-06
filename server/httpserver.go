package server

import (
	"context"
	"log"
	"net/http"
)

type HTTPServer struct {
	tunnelStore *TunnelStorer
	http        http.Server
}

func NewHTTPServer(tunnelStorer *TunnelStorer, port string) *HTTPServer {
	s := &HTTPServer{
		tunnelStore: tunnelStorer,
		http: http.Server{
			Addr: ":" + port,
		},
	}

	s.http.Handler = http.HandlerFunc(s.handleHTTPReq)

	return s
}

func (s *HTTPServer) StartHTTPServer(ctx context.Context, port string) error {
	go func() {
		<-ctx.Done()
		if err := s.http.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	if err := s.http.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("HTTP server ListenAndServe: %v", err)
		return err
	}

	return nil
}

func (s *HTTPServer) handleHTTPReq(w http.ResponseWriter, r *http.Request) {
	log.Printf("New HTTP connection from %s\n", r.RemoteAddr)

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
	s.tunnelStore.Delete(id)
}
