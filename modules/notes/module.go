package notes

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/Illusiard/miniapi/internal/caps"
	"github.com/Illusiard/miniapi/internal/meta"
)

type Module struct{}

func New() *Module { return &Module{} }
func (m *Module) Name() string { return "notes" }

type Note struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type createReq struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type updateReq struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func (m *Module) Register(s caps.Setup) error {
	if s.Store == nil {
		return errConfig("notes module requires Store capability")
	}

	s.Meta.AddEntity(meta.Entity{
		Name:   "Note",
		Table:  "notes",
		Module: m.Name(),
		Fields: []meta.Field{
			{Name: "id", Type: "int", Nullable: false},
			{Name: "title", Type: "string", Nullable: false},
			{Name: "content", Type: "string", Nullable: false},
			{Name: "created_at", Type: "datetime", Nullable: false},
			{Name: "updated_at", Type: "datetime", Nullable: false},
		},
	})

	router := s.Routes.Chi()

	router.Route("/notes", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, req *http.Request) {
			notes, err := listNotes(req.Context(), s)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "list_failed")
				return
			}
			writeJSON(w, http.StatusOK, notes)
		})

		r.Post("/", func(w http.ResponseWriter, req *http.Request) {
			var in createReq
			if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
				writeError(w, http.StatusBadRequest, "invalid_json")
				return
			}
			if in.Title == "" || in.Content == "" {
				writeError(w, http.StatusBadRequest, "title_and_content_required")
				return
			}

			n, err := createNote(req.Context(), s, in.Title, in.Content)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "create_failed")
				return
			}
			writeJSON(w, http.StatusCreated, n)
		})

		r.Get("/{id}", func(w http.ResponseWriter, req *http.Request) {
			id, ok := parseID(w, req)
			if !ok {
				return
			}
			n, found, err := getNote(req.Context(), s, id)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "get_failed")
				return
			}
			if !found {
				writeError(w, http.StatusNotFound, "not_found")
				return
			}
			writeJSON(w, http.StatusOK, n)
		})

		r.Put("/{id}", func(w http.ResponseWriter, req *http.Request) {
			id, ok := parseID(w, req)
			if !ok {
				return
			}
			var in updateReq
			if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
				writeError(w, http.StatusBadRequest, "invalid_json")
				return
			}
			if in.Title == "" || in.Content == "" {
				writeError(w, http.StatusBadRequest, "title_and_content_required")
				return
			}

			n, found, err := updateNote(req.Context(), s, id, in.Title, in.Content)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "update_failed")
				return
			}
			if !found {
				writeError(w, http.StatusNotFound, "not_found")
				return
			}
			writeJSON(w, http.StatusOK, n)
		})

		r.Delete("/{id}", func(w http.ResponseWriter, req *http.Request) {
			id, ok := parseID(w, req)
			if !ok {
				return
			}
			found, err := deleteNote(req.Context(), s, id)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "delete_failed")
				return
			}
			if !found {
				writeError(w, http.StatusNotFound, "not_found")
				return
			}
			w.WriteHeader(http.StatusNoContent)
		})
	})

	return nil
}

func parseID(w http.ResponseWriter, req *http.Request) (int64, bool) {
	idStr := chi.URLParam(req, "id")
	if idStr == "" {
		writeError(w, http.StatusBadRequest, "id_required")
		return 0, false
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid_id")
		return 0, false
	}
	return id, true
}

func listNotes(ctx context.Context, s caps.Setup) ([]Note, error) {
	out := make([]Note, 0, 16)

	err := s.Store.RunInTx(ctx, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `
			select id, title, content, created_at, updated_at
			from notes
			order by id desc
			limit 100
		`)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var n Note
			if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt); err != nil {
				return err
			}
			out = append(out, n)
		}
		return rows.Err()
	})

	return out, err
}

func getNote(ctx context.Context, s caps.Setup, id int64) (Note, bool, error) {
	var n Note
	var found bool

	err := s.Store.RunInTx(ctx, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, `
			select id, title, content, created_at, updated_at
			from notes
			where id = $1
		`, id)

		if err := row.Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt); err != nil {
			if err == pgx.ErrNoRows {
				found = false
				return nil
			}
			return err
		}
		found = true
		return nil
	})

	return n, found, err
}

func createNote(ctx context.Context, s caps.Setup, title, content string) (Note, error) {
	var n Note

	err := s.Store.RunInTx(ctx, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, `
			insert into notes(title, content)
			values($1, $2)
			returning id, title, content, created_at, updated_at
		`, title, content)

		return row.Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt)
	})

	return n, err
}

func updateNote(ctx context.Context, s caps.Setup, id int64, title, content string) (Note, bool, error) {
	var n Note
	var found bool

	err := s.Store.RunInTx(ctx, func(tx pgx.Tx) error {
		row := tx.QueryRow(ctx, `
			update notes
			set title = $2,
			    content = $3,
			    updated_at = now()
			where id = $1
			returning id, title, content, created_at, updated_at
		`, id, title, content)

		if err := row.Scan(&n.ID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt); err != nil {
			if err == pgx.ErrNoRows {
				found = false
				return nil
			}
			return err
		}
		found = true
		return nil
	})

	return n, found, err
}

func deleteNote(ctx context.Context, s caps.Setup, id int64) (bool, error) {
	var rows int64
	err := s.Store.RunInTx(ctx, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `delete from notes where id = $1`, id)
		if err != nil {
			return err
		}
		rows = tag.RowsAffected()
		return nil
	})
	return rows > 0, err
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]string{"error": code})
}

type errConfig string

func (e errConfig) Error() string { return string(e) }
