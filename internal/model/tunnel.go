package model

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"context"
	"time"
)

//counterfeiter:generate -o fake . Tunnel
type Tunnel interface {
	Transfer(ctx context.Context, send string) (string, error)
	Close() error
	Stats() TunnelStats
	SetListener(l TunnelListener)
}

//counterfeiter:generate -o fake . TunnelListener
type TunnelListener interface {
	RemoveTunnel(user User, tun Tunnel)
}

type TunnelStats struct {
	Created       time.Time `json:"created"`
	TransferCount int       `json:"transfer_count"`
	DropCount     int       `json:"drop_count"`
	ReadCount     int       `json:"read_count"`
	PingTxCount   int       `json:"ping_tx_count"`
	PongRxCount   int       `json:"pong_rx_count"`
	PingRxCount   int       `json:"ping_rx_count"`
	PongTxCount   int       `json:"pong_tx_count"`
}
