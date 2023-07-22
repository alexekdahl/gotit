package server

import (
	"context"
	"log"
	"net/http"
)

type HTTPServer struct {
	tunnelStorer TunnelStorer
	http         http.Server
}

func NewHTTPServer(tunnelStorer TunnelStorer, port string) *HTTPServer {
	s := &HTTPServer{
		tunnelStorer: tunnelStorer,
		http: http.Server{
			Addr: ":" + port,
		},
	}

	s.http.Handler = http.HandlerFunc(s.handleHTTPReq)

	return s
}

func (s *HTTPServer) StartHTTPServer(ctx context.Context) error {
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
	tunnelChan, ok := s.tunnelStorer.Get(id)
	if !ok {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	tunnel := Tunnel{
		w:      w,
		donech: make(chan struct{}),
	}
	w.Header().Set("Content-Disposition", `attachment; filename="gotit"`)
	tunnelChan <- tunnel
	<-tunnel.donech
}
