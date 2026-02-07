package caps

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Routes interface {
	Route(pattern string, fn func(r Routes))
	Get(pattern string, h http.HandlerFunc)
	Post(pattern string, h http.HandlerFunc)
}

type chiRoutes struct {
	r chi.Router
}

func NewChiRoutes(r chi.Router) Routes {
	return &chiRoutes{r: r}
}

func (c *chiRoutes) Route(pattern string, fn func(r Routes)) {
	c.r.Route(pattern, func(cr chi.Router) {
		fn(&chiRoutes{r: cr})
	})
}

func (c *chiRoutes) Get(pattern string, h http.HandlerFunc) {
	c.r.Get(pattern, h)
}

func (c *chiRoutes) Post(pattern string, h http.HandlerFunc) {
	c.r.Post(pattern, h)
}
