package server

import (
	"io"
	"net/http"
)

func (s *Server) Invoker(w http.ResponseWriter, r *http.Request) {
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
