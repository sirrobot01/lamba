package handlers

import (
	"github.com/sirrobot01/lamba/internal/server/templates"
	"net/http"
)

func (h *Handler) handleHome(w http.ResponseWriter, r *http.Request) {
	if err := templates.Home(h.ex).Render(r.Context(), w); err != nil {
		return
	}
}
