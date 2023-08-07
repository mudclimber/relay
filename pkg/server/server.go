package server

import (
	"github.com/mudclimber/relay/internal/server"
	"github.com/mudclimber/relay/pkg/handler"
)

func Run(h handler.Handler, opts handler.HandlerOptions) {
  mux := server.MuxServer{}
  mux.Run(h, opts)
}
