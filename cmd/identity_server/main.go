package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"identity/config"
	"identity/server"
)

func main() {
	globalCtx, cancelGlobalCtx := context.WithCancel(context.Background())
	defer cancelGlobalCtx()
	globalConfig := config.ReadConfig()

	serverInstance := server.New(globalCtx, globalConfig)
	serverInstance.RegisterHandlers()
	go serverInstance.Serve()

	// wait for kill signal
	osChan := make(chan os.Signal, 1)
	signal.Notify(osChan, os.Interrupt, syscall.SIGTERM)
	<-osChan
}
