//go:build integration
// +build integration

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/Illusiard/miniapi/internal/caps"
	"github.com/Illusiard/miniapi/internal/meta"
	"github.com/Illusiard/miniapi/internal/migrations"
	"github.com/Illusiard/miniapi/internal/store"
	"github.com/Illusiard/miniapi/modules/notes"
)

type noteDTO struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type createReq struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type updateReq struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func TestNotesCRUDL(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 1) Postgres container
	pg, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("miniapi"),
		tcpostgres.WithUsername("miniapi"),
		tcpostgres.WithPassword("miniapi"),
	)
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	t.Cleanup(func() { _ = pg.Terminate(context.Background()) })

	dbURL, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("conn string: %v", err)
	}

	waitForPostgres(t, ctx, dbURL, 20*time.Second)

	// 2) Migrations
	root := projectRoot(t)
	migPath := filepath.Join(root, "migrations")
	if err := migrations.New(migPath, dbURL).Up(); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	// 3) DB pool + store
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("pgxpool: %v", err)
	}
	t.Cleanup(pool.Close)

	st := store.New(pool)

	// 4) Router + module
	r := chi.NewRouter()
	metaReg := meta.New()

	log := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))

	setup := caps.Setup{
		Routes: caps.NewChiRoutes(r),
		Meta:   metaReg,
		Store:  st,
		Log:    log,
	}

	if err := notes.New().Register(setup); err != nil {
		t.Fatalf("register notes: %v", err)
	}

	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	// 5) CREATE
	created := mustDoJSON[noteDTO](t, http.MethodPost, srv.URL+"/notes", createReq{
		Title:   "Hello",
		Content: "World",
	}, http.StatusCreated)

	if created.ID <= 0 {
		t.Fatalf("expected id > 0, got %d", created.ID)
	}
	if created.Title != "Hello" || created.Content != "World" {
		t.Fatalf("unexpected created note: %+v", created)
	}

	// 6) READ one
	got := mustDoJSON[noteDTO](t, http.MethodGet, urlf(srv.URL+"/notes/%d", created.ID), nil, http.StatusOK)
	if got.ID != created.ID {
		t.Fatalf("expected same id, got=%d want=%d", got.ID, created.ID)
	}

	// 7) LIST
	list := mustDoJSON[[]noteDTO](t, http.MethodGet, srv.URL+"/notes", nil, http.StatusOK)
	if len(list) == 0 {
		t.Fatalf("expected list not empty")
	}

	// 8) UPDATE
	updated := mustDoJSON[noteDTO](t, http.MethodPut, urlf(srv.URL+"/notes/%d", created.ID), updateReq{
		Title:   "Hello2",
		Content: "World2",
	}, http.StatusOK)
	if updated.Title != "Hello2" || updated.Content != "World2" {
		t.Fatalf("unexpected updated note: %+v", updated)
	}

	// 9) DELETE
	mustDoNoBody(t, http.MethodDelete, urlf(srv.URL+"/notes/%d", created.ID), http.StatusNoContent)

	// 10) Ensure deleted
	mustDoJSON[map[string]string](t, http.MethodGet, urlf(srv.URL+"/notes/%d", created.ID), nil, http.StatusNotFound)

	{
		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/notes", bytes.NewReader([]byte(`{"title":`)))
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("invalid json request: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			raw, _ := io.ReadAll(resp.Body)
			t.Fatalf("invalid json: status got=%d want=%d, body=%s", resp.StatusCode, http.StatusBadRequest, string(raw))
		}
		var out map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&out)
		if out["error"] != "invalid_json" {
			t.Fatalf("invalid json: error got=%q want=%q", out["error"], "invalid_json")
		}
	}

	// empty fields
	mustErrorCode(t, http.MethodPost, srv.URL+"/notes", createReq{Title: "", Content: ""}, http.StatusBadRequest, "title_and_content_required")

	// invalid id (non-numeric)
	mustErrorCode(t, http.MethodGet, srv.URL+"/notes/abc", nil, http.StatusBadRequest, "invalid_id")

	// invalid id (zero)
	mustErrorCode(t, http.MethodGet, srv.URL+"/notes/0", nil, http.StatusBadRequest, "invalid_id")

	// not found
	mustErrorCode(t, http.MethodGet, srv.URL+"/notes/99999999", nil, http.StatusNotFound, "not_found")
}

