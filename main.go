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
	"sync"
	"syscall"
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
		clusterMap[cc.Name] = core.NewCluster(cc.Name, cc.Endpoints, cc.ConnectTimeout.Duration, logger)
	}

	//Create context that cancels on SIGINT/SIGTERM
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// WaitGroup to track all listener goroutines
	var wg sync.WaitGroup

	//Start Listeners
	for _, lc := range cfg.Listeners {
		cluster, ok := clusterMap[lc.Cluster]
		if !ok {
			logger.Error("Listener references unknown cluster",
				slog.String("Listener", lc.Name),
				slog.String("Cluster", lc.Cluster))
			os.Exit(1)
		}
		//Chain filters

		chainFilters := []api.NetworkFilter{
			filters.NewLogger(logger),
			filters.NewTCPProxy(cluster, logger),
		}

		listener := core.NewListener(lc, chainFilters, logger)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := listener.Start(ctx); err != nil {
				logger.Error("Listener error", slog.String("error", err.Error()))
			}
		}()

	}

	//Block until shutdown signal
	<-ctx.Done()
	logger.Info("Shutting down Conduit")

	// Wait for all listeners to finish draining connections
	wg.Wait()
	logger.Info("All listeners stopped. Goodbye!")
}
