package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/sirrobot01/lamba/server/templates"
	"net/http"
)

func (s *Server) Routes(r *chi.Mux) {
	r.Get("/", s.handleHome)
	r.Post("/invoke", s.Invoker)
	r.Route("/functions", func(r chi.Router) {
		r.Get("/", s.handleFunctionsList)
		r.Post("/", s.handleFunctionsCreate)
		r.Delete("/{name}", s.handleFunctionsDelete)
	})

	r.Route("/events", func(r chi.Router) {
		r.Get("/", s.handleEventList)
		r.Post("/", s.handleEventCreate)
		r.Delete("/{id}", s.handleEventDelete)
	})

}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	if err := templates.Home(s.ex).Render(r.Context(), w); err != nil {
		return
	}
}
