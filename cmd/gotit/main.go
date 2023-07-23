package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/AlexEkdahl/gotit/server"
	"github.com/AlexEkdahl/gotit/utils/logger"
)

func main() {
	httpPort := flag.String("httpport", "8080", "Port to the http server")
	sshPort := flag.String("sshport", "2222", "Port to the ssh server")
	env := flag.String("env", "LOCAL", "Environment")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p, _ := filepath.Abs("loggs.txt")
	c := logger.Config{
		Env:  *env,
		Path: p,
	}

	logger, err := logger.NewLogger(c)
	if err != nil {
		log.Printf("Could not instantiate log %s", err.Error())
		cancel()
	}
	defer logger.Close()

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigch
		cancel()
	}()

	tunnelStorer := server.NewTunnel()
	httpServer := server.NewHTTPServer(tunnelStorer, logger, *httpPort)
	sshServer, err := server.NewSSHServer(tunnelStorer, logger, *sshPort)
	if err != nil {
		logger.Error(err)
		cancel()
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := httpServer.StartHTTPServer(ctx); err != nil {
			logger.Error(err)
			cancel()

		}
	}()

	go func() {
		defer wg.Done()
		if err := sshServer.StartSSHServer(ctx); err != nil {
			logger.Error(err)
			cancel()
		}
	}()

	logger.Info("Environment: %s", *env)
	logger.Info("Servers up and running on sshport: %s, httport: %s", *sshPort, *httpPort)
	wg.Wait()
}
