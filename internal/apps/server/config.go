package server

import "github.com/dlomanov/mon/internal/apps/server/handlers"

type Config struct {
	handlers.Config
	Addr string
}
