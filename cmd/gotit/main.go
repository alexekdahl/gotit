package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/AlexEkdahl/gotit/server"
)

func main() {
	httpPort := flag.String("httpport", "8080", "Port to the http server")
	sshPort := flag.String("sshport", "2222", "Port to the ssh server")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigch
		cancel()
	}()

	tunnelStorer := server.NewTunnel()
	httpServer := server.NewHTTPServer(tunnelStorer, *httpPort)
	sshServer, err := server.NewSSHServer(tunnelStorer, *sshPort)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := httpServer.StartHTTPServer(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := sshServer.StartSSHServer(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	wg.Wait()
}
