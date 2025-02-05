package server

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
	"github.com/sirrobot01/lamba/pkg/executor"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type Server struct {
	ex   *executor.Executor
	port string
}

func NewServer(ex *executor.Executor, port string) *Server {
	return &Server{
		ex:   ex,
		port: port,
	}
}

func (s *Server) Start() error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	srv := &http.Server{
		Addr:    ":" + s.port,
		Handler: r,
	}
	s.Routes(r)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Info().Msgf("Starting server on %s", srv.Addr)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Info().Msgf("Error starting server: %v\n", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Info().Msgf("Shutting down gracefully...")
	return srv.Shutdown(context.Background())
}
