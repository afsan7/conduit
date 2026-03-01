package main

import (
	"conduit/pkg/api"
	"conduit/pkg/config"
	"conduit/pkg/core"
	"conduit/pkg/filters"
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	configPath := flag.String("config", "config.json", "path to the config file")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Load Config

	cfg, err := config.LoadFromFile(*configPath)

	if err != nil {
		logger.Error("Failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Create Clusters
	clusterMap := make(map[string]*core.Cluster)
	for _, cc := range cfg.Clusters {
		timeout, _ := time.ParseDuration(cc.ConnectTimeout)
		clusterMap[cc.Name] = core.NewCluster(cc.Name, cc.Endpoints, timeout, logger)
	}

	//Create context that cancels on SIGINT/SIGTERM
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	//Start Listeners
	for _, lc := range cfg.Listeners {

		//Chain filters

		chainFilters := []api.NetworkFilter{
			filters.NewLogger(logger),
			filters.NewTCPProxy(clusterMap["backend"], logger),
		}

		listener := core.NewListener(lc, chainFilters, logger)
		go func() {
			if err := listener.Start(ctx); err != nil {
				logger.Error("Listener error", slog.String("error", err.Error()))
			}
		}()

	}

	//Block until shutdown signal
	<-ctx.Done()
	logger.Info("Shutting down Conduit")
}
