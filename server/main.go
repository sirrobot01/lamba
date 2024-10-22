package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirrobot01/lamba/pkg/executor"
	"log"
	"net/http"
)

type Server struct {
	ex *executor.Executor
}

func NewServer(ex *executor.Executor) *Server {
	return &Server{
		ex: ex,
	}
}

func (s *Server) Start() error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	s.Routes(r)

	log.Println("Server starting on :8080")

	if err := http.ListenAndServe(":8080", r); err != nil {
		return err
	}
	return nil
}
