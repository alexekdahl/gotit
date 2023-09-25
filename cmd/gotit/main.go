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

	"github.com/AlexEkdahl/gotit/pkg/http"
	"github.com/AlexEkdahl/gotit/pkg/pipe"
	"github.com/AlexEkdahl/gotit/pkg/ssh"
	"github.com/AlexEkdahl/gotit/pkg/util"
)

func main() {
	httpPort := flag.String("httpport", "8080", "Port to the http server")
	sshPort := flag.String("sshport", "2222", "Port to the ssh server")
	env := flag.String("env", "LOCAL", "Environment")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p, _ := filepath.Abs("loggs.txt")
	c := util.Config{
		Env:  *env,
		Path: p,
	}

	logger, err := util.NewSimpleLogger(c)
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

	tunnelStorer := pipe.NewTunnelStore()
	httpServer := http.NewServer(tunnelStorer, logger, *httpPort)
	sshServer, err := ssh.NewServer(tunnelStorer, logger, *sshPort)
	if err != nil {
		logger.Error(err)
		cancel()
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := httpServer.StartServer(ctx); err != nil {
			logger.Error(err)
			cancel()

		}
	}()

	go func() {
		defer wg.Done()
		if err := sshServer.StartServer(ctx); err != nil {
			logger.Error(err)
			cancel()
		}
	}()

	logger.Info("Environment: %s", *env)
	logger.Info("Servers up and running on sshport: %s, httport: %s", *sshPort, *httpPort)
	wg.Wait()
}
