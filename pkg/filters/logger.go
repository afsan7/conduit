package filters

import (
	"conduit/pkg/api"
	"context"
	"log/slog"
	"net"
	"time"
)

type Logger struct {
	logger *slog.Logger
}

func NewLogger(logger *slog.Logger) *Logger {
	if logger == nil {
		logger = slog.Default()
	}

	return &Logger{logger: logger}
}

func (l *Logger) Name() string {
	return "conduit.filters.logger"
}

func (l *Logger) OnNewConnection(ctx context.Context, conn net.Conn) api.FilterStatus {
	remoteAddr := conn.RemoteAddr()
	localAddr := conn.LocalAddr()

	l.logger.InfoContext(ctx, "New Connection",
		slog.String("remote_addr", remoteAddr.String()),
		slog.String("local_addr", localAddr.String()),
		slog.String("network", remoteAddr.Network()),
		slog.Time("connected_at", time.Now()),
	)

	return api.Continue

}
