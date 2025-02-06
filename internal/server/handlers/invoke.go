package handlers

import (
	"io"
	"net/http"
)

func (h *Handler) Invoker(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	switch contentType {
	case "application/json":
		w.Header().Set("Content-Type", "application/json")
	case "text/plain":
		w.Header().Set("Content-Type", "text/plain")
	default:
		w.Header().Set("Content-Type", "application/json")
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "function id not provided", http.StatusBadRequest)
		return
	}
	body, _ := io.ReadAll(r.Body)
	result, err := h.ex.Execute("http", id, string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write([]byte(result))
}
