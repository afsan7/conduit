package api

import (
	"context"
	"net"
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
}

type ClusterConfig struct {
	Name           string   `json:"name"`
	Endpoints      []string `json:"endpoints"`
	ConnectTimeout string   `json:"connect_timeout"`
}

type Config struct {
	Listeners []ListenerConfig `json:"listeners"`
	Clusters  []ClusterConfig  `json:"clusters"`
}
