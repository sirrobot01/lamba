package handlers

import (
	"github.com/sirrobot01/lamba/internal/server/components"
	"net/http"
)

func (h *Handler) handleEventCreate(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	payload := r.FormValue("payload")
	_, err := h.ex.Execute("http", id, payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = components.EventList(h.ex).Render(r.Context(), w)
}

func (h *Handler) handleEventList(w http.ResponseWriter, r *http.Request) {
	if err := components.EventList(h.ex).Render(r.Context(), w); err != nil {
		return
	}
}

func (h *Handler) handleEventDelete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	h.ex.EventManager.Remove(id)
	_ = components.EventList(h.ex).Render(r.Context(), w)
}
