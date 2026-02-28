package core

import (
	"conduit/pkg/api"
	"context"
	"log/slog"
	"net"
	"sync"
)

type Listener struct {
	name    string
	address string
	ln      net.Listener
	filters []api.NetworkFilter
	logger  *slog.Logger
	wg      sync.WaitGroup
}

func NewListener(cfg api.ListenerConfig, filters []api.NetworkFilter, logger *slog.Logger) *Listener {
	if logger == nil {
		logger = slog.Default()
	}

	return &Listener{
		name:    cfg.Name,
		address: cfg.Address,
		filters: filters,
		logger:  logger,
	}
}

func (l *Listener) Start(ctx context.Context) error {
	var err error
	l.ln, err = net.Listen("tcp", l.address)
	if err != nil {
		return err
	}

	l.logger.Info("Listener Started", slog.String("name", l.name), slog.String("address", l.address))

	go func() {
		<-ctx.Done()
		l.ln.Close()
	}()

	for {
		conn, err := l.ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			l.logger.Error("Failed to Accept connection", slog.String("listener", l.name), slog.String("error", err.Error()))
			continue
		}

		l.wg.Add(1)
		go l.handleConnection(ctx, conn)

	}

	l.wg.Wait()
	l.logger.Info("Listener stopped", slog.String("name", l.name))
	return nil
}

func (l *Listener) handleConnection(ctx context.Context, conn net.Conn) {
	defer l.wg.Done()
	defer conn.Close()

	for _, f := range l.filters {
		status := f.OnNewConnection(ctx, conn)
		if status == api.Stop {
			break
		}
	}
}
