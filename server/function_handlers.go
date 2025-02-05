package server

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/sirrobot01/lamba/common"
	"github.com/sirrobot01/lamba/server/components"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

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
		log.Info().Msgf("No file uploaded: %v", err)
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

	rtnPath := filepath.Join("assets/functions", _runtime)
	if err := os.MkdirAll(rtnPath, 0755); err != nil {
		http.Error(w, "Failed to create function directory", http.StatusInternalServerError)
		return
	}
	err = common.ExtractZip(file, h.Size, rtnPath)
	functionDir, _ := filepath.Abs(filepath.Join(rtnPath, name))
	if err != nil {
		// Clean up directory if extraction fails
		os.RemoveAll(functionDir)
		http.Error(w, fmt.Sprintf("Failed to extract zip: %v", err), http.StatusInternalServerError)
		return
	}

	if err := s.ex.CreateFunction(name, _runtime, handler, timeout, functionDir, preExec); err != nil {
		log.Info().Err(err).Msgf("Error creating function")
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
