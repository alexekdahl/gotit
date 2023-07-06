package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/AlexEkdahl/gotit/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigch
		cancel()
	}()

	var wg sync.WaitGroup
	wg.Add(2)

	tunnelStore := server.NewTunnel()
	httpServer := server.NewHTTPServer(tunnelStore, "8080")
	sshServer := server.NewSSHServer(tunnelStore, "2222")

	go func() {
		defer wg.Done()
		if err := httpServer.StartHTTPServer(ctx, "8080"); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := sshServer.StartSSHServer(ctx, "2222"); err != nil {
			log.Fatal(err)
		}
	}()

	wg.Wait()
}
