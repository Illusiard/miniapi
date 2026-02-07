package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Illusiard/miniapi/internal/caps"
	"github.com/Illusiard/miniapi/internal/config"
	"github.com/Illusiard/miniapi/internal/db"
	"github.com/Illusiard/miniapi/internal/httpserver"
	"github.com/Illusiard/miniapi/internal/meta"
	"github.com/Illusiard/miniapi/internal/modules"
	"github.com/Illusiard/miniapi/internal/store"
	"github.com/Illusiard/miniapi/modules/ping"
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

	pgStore := store.New(a.db)

	readyFn := func(ctx context.Context) error {
		return pgStore.Ping(ctx)
	}

	metaReg := meta.New()

	registerFn := func(r chi.Router) {
		r.Get("/meta/entities", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(metaReg.Entities())
		})
		r.Get("/meta/modules", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(metaReg.Modules())
		})


		specs := []modules.Spec{
			{Module: ping.New(), WithStore: false},
		}

		for _, spec := range specs {
			m := spec.Module
			slog.Info("registering module", "module", m.Name())

			setup := caps.Setup{
				Routes: caps.NewChiRoutes(r),
				Meta:   metaReg,
				Log:    slog.Default(),
			}
			if spec.WithStore {
				setup.Store = pgStore
			}

			if err := m.Register(setup); err != nil {
				panic(fmt.Errorf("module %s register: %w", m.Name(), err))
			}
			metaReg.AddModule(meta.Module{
				Name:      m.Name(),
				WithStore: spec.WithStore,
			})
		}
	}

	a.server = httpserver.New(a.cfg.HTTPAddr, readyFn, registerFn)

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
