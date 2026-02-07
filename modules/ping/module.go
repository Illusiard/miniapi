package ping

import (
	"net/http"
	"time"

	"github.com/Illusiard/miniapi/internal/caps"
	"github.com/Illusiard/miniapi/internal/meta"
)

type Module struct{}

func New() *Module { return &Module{} }

func (m *Module) Name() string { return "ping" }

func (m *Module) Register(s caps.Setup) error {
	s.Meta.AddEntity(meta.Entity{
		Name:   "Ping",
		Table:  "",
		Module: m.Name(),
		Fields: []meta.Field{
			{Name: "ts", Type: "datetime", Nullable: false},
			{Name: "message", Type: "string", Nullable: false},
		},
	})

	s.Routes.Route("/ping", func(r caps.Routes) {
		r.Get("/", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_, _ = w.Write([]byte(`{"message":"pong","ts":"` + time.Now().UTC().Format(time.RFC3339Nano) + `"}`))
		})
	})

	return nil
}
