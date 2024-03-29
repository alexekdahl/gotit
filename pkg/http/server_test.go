package http

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/AlexEkdahl/gotit/pkg/pipe"
	"github.com/AlexEkdahl/gotit/pkg/util"
)

func TestServer(t *testing.T) {
	// Start the HTTP server
	l := util.NewMockLogger()
	tunnelStore := pipe.NewTunnelStore()
	server := NewServer(tunnelStore, l, "8080")

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.StartServer(context.Background())
	}()

	// Wait for the server to start
	time.Sleep(time.Second)

	// Check for server start error
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Failed to start HTTP server: %s", err)
		}
	default:
	}

	// Make a GET request
	resp, err := http.Get("http://localhost:8080/?id=testID")
	if err != nil {
		t.Fatalf("Failed to make GET request: %s", err)
	}
	defer resp.Body.Close()

	// Check the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %s", err)
	}

	if resp.StatusCode != http.StatusNotFound || string(body) != "Not Found\n" {
		t.Errorf("Expected status 404 Not Found, got %d %s", resp.StatusCode, body)
	}
}
