package registry

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Illusiard/miniapi/internal/meta"
)

type Registry struct {
	DB      *pgxpool.Pool
	Logger  *slog.Logger
	Meta    *meta.Registry
	RootMux Router
}

type Router interface {
	Route(pattern string, fn func(r Router))
	Get(pattern string, handler func(w ResponseWriter, r *Request))
}

type ResponseWriter interface {
	Header() map[string][]string
	Write([]byte) (int, error)
	WriteHeader(statusCode int)
}

type Request struct {
	Method string
	URL    string
}
