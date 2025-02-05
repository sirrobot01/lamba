package server

import (
	"fmt"
	"io"
	"net/http"
)

func (s *Server) Invoker(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	fmt.Print(contentType)
	switch contentType {
	case "application/json":
		w.Header().Set("Content-Type", "application/json")
	case "text/plain":
		w.Header().Set("Content-Type", "text/plain")
	default:
		w.Header().Set("Content-Type", "application/json")
	}
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "function name not provided", http.StatusBadRequest)
		return
	}
	body, _ := io.ReadAll(r.Body)
	result, err := s.ex.Execute("http", name, string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write([]byte(result))
}
