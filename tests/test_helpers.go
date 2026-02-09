//go:build integration
// +build integration

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func projectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("project root not found (go.mod)")
		}
		dir = parent
	}
}

func mustDoJSON[T any](t *testing.T, method, url string, body any, wantStatus int) T {
	t.Helper()

	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("json marshal: %v", err)
		}
		r = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, r)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != wantStatus {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("%s %s: status got=%d want=%d, body=%s",
			method, url, resp.StatusCode, wantStatus, string(raw))
	}

	var out T
	if wantStatus == http.StatusNoContent {
		return out
	}

	dec := json.NewDecoder(resp.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&out); err != nil {
		t.Fatalf("decode json: %v", err)
	}

	return out
}

func mustDoNoBody(t *testing.T, method, url string, wantStatus int) {
	t.Helper()

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != wantStatus {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("%s %s: status got=%d want=%d, body=%s",
			method, url, resp.StatusCode, wantStatus, string(raw))
	}
}

func urlf(base string, parts ...any) string {
	return fmt.Sprintf(base, parts...)
}

func mustErrorCode(t *testing.T, method, url string, body any, wantStatus int, wantCode string) {
	t.Helper()

	resp := mustDoJSON[map[string]string](t, method, url, body, wantStatus)
	code := resp["error"]
	if code != wantCode {
		t.Fatalf("%s %s: error code got=%q want=%q", method, url, code, wantCode)
	}
}

func waitForPostgres(t *testing.T, ctx context.Context, dbURL string, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)

	for {
		if time.Now().After(deadline) {
			t.Fatalf("postgres not ready after %s", timeout)
		}

		db, err := sql.Open("pgx", dbURL)
		if err == nil {
			pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			err = db.PingContext(pingCtx)
			cancel()
			_ = db.Close()
			if err == nil {
				return
			}
		}

		time.Sleep(300 * time.Millisecond)
	}
}
