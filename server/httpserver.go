package server

import (
	"context"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/AlexEkdahl/gotit/utils/logger"
)

type templateRenderer struct {
	template *template.Template
}

func (t *templateRenderer) render(w io.Writer, name string, data interface{}) error {
	return t.template.ExecuteTemplate(w, name, data)
}

type HTTPServer struct {
	tunnelStorer     TunnelStorer
	templateRenderer *templateRenderer
	http             http.Server
	logger           logger.Logger
}

func NewHTTPServer(tunnelStorer TunnelStorer, logger logger.Logger, port string) *HTTPServer {
	tmplPath := filepath.Join("www/templates/index.html")
	t, err := template.ParseFiles(tmplPath)
	if err != nil {
		return nil
	}

	s := &HTTPServer{
		tunnelStorer: tunnelStorer,
		templateRenderer: &templateRenderer{
			template: t,
		},
		http:   http.Server{Addr: ":" + port},
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
	s.logger.Debug("URL: " + r.URL.Path)
	if r.URL.Path == "/" {
		if err := s.templateRenderer.render(w, "index.html", nil); err != nil {
			s.logger.Error(err)
			return
		}

		return
	}

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
