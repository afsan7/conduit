package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type FilterStatus int

const (
	Continue FilterStatus = iota
	Stop
)

type NetworkFilter interface {
	Name() string
	OnNewConnection(ctx context.Context, conn net.Conn) FilterStatus
}

// -- Configuration Structs --

type ListenerConfig struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Cluster string `json:"cluster"`
}

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw == "" {
		d.Duration = 0
		return nil
	}

	parsed, err := time.ParseDuration(raw)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", raw, err)
	}

	d.Duration = parsed
	return nil
}

type ClusterConfig struct {
	Name           string   `json:"name"`
	Endpoints      []string `json:"endpoints"`
	ConnectTimeout Duration `json:"connect_timeout"`
}

type Config struct {
	Listeners []ListenerConfig `json:"listeners"`
	Clusters  []ClusterConfig  `json:"clusters"`
}
