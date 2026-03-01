package filters

import (
	"conduit/pkg/api"
	"conduit/pkg/core"
	"context"
	"io"
	"log/slog"
	"net"
)

type TCPProxy struct {
	cluster *core.Cluster
	logger  *slog.Logger
}

func NewTCPProxy(cluster *core.Cluster, logger *slog.Logger) *TCPProxy {
	if logger == nil {
		logger = slog.Default()
	}

	return &TCPProxy{cluster: cluster, logger: logger}
}
func (t *TCPProxy) Name() string {
	return "conduit.filters.tcp_proxy"
}

func (t *TCPProxy) OnNewConnection(ctx context.Context, conn net.Conn) api.FilterStatus {
	//Dial upstream
	upstream, err := t.cluster.Dial(ctx)
	if err != nil {
		t.logger.ErrorContext(ctx, "Failed to connect upstream", slog.String("error", err.Error()))
		return api.Stop
	}

	defer upstream.Close()

	errCh := make(chan error, 2)

	// Upstream -> client
	go func() {
		_, err := io.Copy(conn, upstream)
		errCh <- err
	}()

	// Client -> upstream

	go func() {
		_, err := io.Copy(upstream, conn)
		errCh <- err
	}()

	<-errCh

	return api.Stop
}
