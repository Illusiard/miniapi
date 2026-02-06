package app

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Illusiard/miniapi/internal/config"
	"github.com/Illusiard/miniapi/internal/httpserver"
)

type App struct {
	cfg    config.Config
	server *httpserver.Server
}

func New(cfg config.Config) *App {
	return &App{
		cfg:    cfg,
		server: httpserver.New(cfg.HTTPAddr),
	}
}

func (a *App) Start(ctx context.Context) error {
	slog.Info("starting http server", "addr", a.cfg.HTTPAddr)
	return a.server.Start(ctx)
}

func (a *App) Stop(ctx context.Context) error {
	slog.Info("stopping http server")
	return a.server.Stop(ctx)
}

// чтобы не ругался линтер на импорт net/http, когда расширим:
var _ = http.StatusOK
