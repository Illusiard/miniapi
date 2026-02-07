package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Illusiard/miniapi/internal/config"
	"github.com/Illusiard/miniapi/internal/db"
	"github.com/Illusiard/miniapi/internal/httpserver"
)

type App struct {
	cfg    config.Config

	db     *pgxpool.Pool
	server *httpserver.Server
}

func New(cfg config.Config) *App {
	return &App{cfg: cfg}
}

func (a *App) Start(ctx context.Context) error {
	slog.Info("starting http server", "addr", a.cfg.HTTPAddr)
	connCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err := db.Connect(connCtx, a.cfg.DatabaseURL)
	if err != nil {
		return err
	}
	a.db = pool

	readyFn := func(ctx context.Context) error {
		return a.db.Ping(ctx)
	}

	a.server = httpserver.New(a.cfg.HTTPAddr, readyFn)

	slog.Info("starting http server", "addr", a.cfg.HTTPAddr)
	if err := a.server.Start(ctx); err != nil {
		return fmt.Errorf("http server: %w", err)
	}

	return nil
}

func (a *App) Stop(ctx context.Context) error {
	slog.Info("stopping http server")
	if a.server != nil {
		_ = a.server.Stop(ctx)
	}
	if a.db != nil {
		a.db.Close()
	}
	return nil
}

var _ = http.StatusOK

