package handlers

import (
	"errors"
	"io"
	"mime/multipart"
	"path/filepath"
)

type functionCreateRequest struct {
	Name       string
	Runtime    string
	Handler    string
	FileHeader *multipart.FileHeader
	File       io.ReaderAt
	PreExec    string
	Timeout    int
}

func (h *Handler) validateFunctionCreate(input *functionCreateRequest) error {
	if input.Name == "" {
		return errors.New("name is required")
	}

	if fn := h.ex.FunctionRegistry.GetByName(input.Name, input.Runtime); fn != nil {
		return errors.New("function name already exists")
	}

	if input.Runtime == "" {
		return errors.New("runtime is required")
	}

	if _, exists := h.ex.RuntimeManager.Get(input.Runtime); !exists {
		return errors.New("runtime does not exist")
	}

	if input.FileHeader == nil {
		return errors.New("file is required")
	}

	if filepath.Ext(input.FileHeader.Filename) != ".zip" {
		return errors.New("only ZIP files are allowed")
	}

	return nil
}
