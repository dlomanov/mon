package server

import "github.com/dlomanov/mon/internal/apps/server/container"

type Config struct {
	container.Config
	Addr string
}
