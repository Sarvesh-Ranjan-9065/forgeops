// Command forge-webhook runs the ForgeOps webhook server.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sarvesh-ranjan-9065/forgeops/internal/config"
	"github.com/sarvesh-ranjan-9065/forgeops/internal/log"
	"github.com/sarvesh-ranjan-9065/forgeops/internal/server"
)

func main() {
	logger := log.New("info")

	secret, err := config.WebhookSecret()
	if err != nil {
		logger.Error("config error", "error", err)
		os.Exit(1)
	}

	router := server.NewRouter(server.RouterDeps{
		WebhookSecret: []byte(secret),
	})

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv := server.New(":8080", router, logger)
	if err := srv.Run(ctx); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}
