package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/sirrobot01/lamba/internal/server/components"
	"net/http"
	"strconv"
)

func (h *Handler) handleFunctionsList(w http.ResponseWriter, r *http.Request) {
	if err := components.FunctionList(h.ex).Render(r.Context(), w); err != nil {
		return
	}
}

func (h *Handler) handleFunctionsCreate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	file, fh, err := r.FormFile("file")
	if err != nil {
		// Handle case where no file was uploaded
		log.Info().Msgf("No file uploaded: %v", err)
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	timeout, _ := strconv.Atoi(r.FormValue("timeout"))
	if timeout == 0 {
		timeout = 30 // default timeout
	}

	input := &functionCreateRequest{
		Name:       r.FormValue("name"),
		Runtime:    r.FormValue("runtime"),
		Handler:    r.FormValue("handler"),
		FileHeader: fh,
		File:       file,
		PreExec:    r.FormValue("preExec"),
		Timeout:    timeout,
	}

	if err := h.validateFunctionCreate(input); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.ex.CreateFunction(input.Name, input.Runtime, input.Handler, input.Timeout, input.PreExec, file, fh.Size); err != nil {
		log.Info().Err(err).Msgf("Error creating function")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = components.FunctionList(h.ex).Render(r.Context(), w)
}

func (h *Handler) handleFunctionsDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.ex.DeleteFunction(id); err != nil {
		log.Info().Err(err).Msgf("Error deleting function")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = components.FunctionList(h.ex).Render(r.Context(), w)
}
