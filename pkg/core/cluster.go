package core

import (
	"context"
	"log/slog"
	"net"
	"sync/atomic"
	"time"
)

// Cluster represents a pool of upstream backend endpoints.
// It handles load balancing and dialing upstream connections.
type Cluster struct {
	name           string
	endpoints      []string
	connectTimeout time.Duration
	nextIdx        atomic.Uint64
	logger         *slog.Logger
}

func NewCluster(name string, endpoints []string, connectTimeout time.Duration, logger *slog.Logger) *Cluster {
	if logger == nil {
		logger = slog.Default()
	}
	if connectTimeout == 0 {
		connectTimeout = 5 * time.Second
	}
	return &Cluster{
		name:           name,
		endpoints:      endpoints,
		connectTimeout: connectTimeout,
		logger:         logger,
	}
}

// NextEndpoint returns the next upstream endpoint using round-robin.
func (c *Cluster) NextEndpoint() string {
	idx := c.nextIdx.Add(1)
	return c.endpoints[idx%uint64(len(c.endpoints))]
}

// Dial connects to the next available upstream endpoint.
func (c *Cluster) Dial(ctx context.Context) (net.Conn, error) {
	endpoint := c.NextEndpoint()

	c.logger.DebugContext(ctx, "dialing upstream",
		slog.String("cluster", c.name),
		slog.String("endpoint", endpoint),
	)

	conn, err := net.DialTimeout("tcp", endpoint, c.connectTimeout)
	if err != nil {
		c.logger.ErrorContext(ctx, "upstream dial failed",
			slog.String("cluster", c.name),
			slog.String("endpoint", endpoint),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	return conn, nil
}

func (c *Cluster) Name() string {
	return c.name
}
