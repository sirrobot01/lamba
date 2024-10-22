package server

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/sirrobot01/lamba/common"
	"github.com/sirrobot01/lamba/server/components"
	"github.com/sirrobot01/lamba/server/templates"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
	})

}

func (s *Server) Invoker(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("function")
	body, _ := io.ReadAll(r.Body)
	result, err := s.ex.Execute("http", name, body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(result)
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	if err := templates.Home(s.ex).Render(r.Context(), w); err != nil {
		return
	}
}

func (s *Server) handleFunctionsList(w http.ResponseWriter, r *http.Request) {
	if err := components.FunctionList(s.ex).Render(r.Context(), w); err != nil {
		return
	}
}

func (s *Server) handleFunctionsCreate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	file, h, err := r.FormFile("file")
	if err != nil {
		// Handle case where no file was uploaded
		log.Printf("No file uploaded: %v", err)
	}
	defer file.Close()
	if filepath.Ext(h.Filename) != ".zip" {
		http.Error(w, "Only ZIP files are allowed", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	_runtime := r.FormValue("runtime")
	handler := r.FormValue("handler")

	preExec := r.FormValue("preExec")
	timeout, err := strconv.Atoi(r.FormValue("timeout"))
	if err != nil {
		timeout = 30
	}

	functionDir := filepath.Join("assets/functions", _runtime, name)
	if err := os.MkdirAll(functionDir, 0755); err != nil {
		http.Error(w, "Failed to create function directory", http.StatusInternalServerError)
		return
	}
	err = common.ExtractZip(file, h.Size, functionDir)
	if err != nil {
		// Clean up directory if extraction fails
		os.RemoveAll(functionDir)
		http.Error(w, fmt.Sprintf("Failed to extract zip: %v", err), http.StatusInternalServerError)
		return
	}

	functionPath, _ := filepath.Abs(functionDir)

	if err := s.ex.CreateFunction(name, _runtime, handler, timeout, functionPath, preExec); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = components.FunctionList(s.ex).Render(r.Context(), w)
}

func (s *Server) handleFunctionsDelete(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if err := s.ex.DeleteFunction(name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = components.FunctionList(s.ex).Render(r.Context(), w)
}

func (s *Server) handleEventCreate(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	payload := []byte(r.FormValue("payload"))
	_, err := s.ex.Execute("http", name, payload)
	if err != nil {
		log.Printf("Error executing function: %v", err)
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
