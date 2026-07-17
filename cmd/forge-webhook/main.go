// Command forge-webhook runs the ForgeOps webhook server.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sarvesh-ranjan-9065/forgeops/internal/config"
	"github.com/sarvesh-ranjan-9065/forgeops/internal/gh"
	"github.com/sarvesh-ranjan-9065/forgeops/internal/log"
	"github.com/sarvesh-ranjan-9065/forgeops/internal/preview"
	"github.com/sarvesh-ranjan-9065/forgeops/internal/server"
	"github.com/sarvesh-ranjan-9065/forgeops/k8s"
)

func main() {
	logger := log.New("info")

	secret, err := config.WebhookSecret()
	if err != nil {
		logger.Error("config error", "error", err)
		os.Exit(1)
	}

	cfg, err := config.FromEnv()
	if err != nil {
		logger.Error("config error", "error", err)
		os.Exit(1)
	}

	clientset, err := k8s.NewClientset()
	if err != nil {
		logger.Error("kube client error", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	prov := &preview.Provisioner{
		Clientset: clientset,
		GitHub:    gh.NewClient(ctx, cfg.GitHubToken),
		Logger:    logger,
		Registry:  "localhost:5000",
	}
	worker := preview.NewWorker(prov, 64)
	go worker.Run(ctx)

	router := server.NewRouter(server.RouterDeps{
		WebhookSecret:  []byte(secret),
		WebhookHandler: preview.NewWebhookHandler(worker, logger),
	})

	srv := server.New(":8080", router, logger)
	if err := srv.Run(ctx); err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}
}
