package server

import (
	"github.com/sirrobot01/lamba/server/components"
	"net/http"
)

func (s *Server) handleEventCreate(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	payload := r.FormValue("payload")
	_, err := s.ex.Execute("http", name, payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = components.EventList(s.ex).Render(r.Context(), w)
}

func (s *Server) handleEventList(w http.ResponseWriter, r *http.Request) {
	if err := components.EventList(s.ex).Render(r.Context(), w); err != nil {
		return
	}
}

func (s *Server) handleEventDelete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	s.ex.EventManager.Remove(id)
	_ = components.EventList(s.ex).Render(r.Context(), w)
}
