package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirrobot01/lamba/pkg/executor"
	"log"
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

	log.Printf("Starting server on %s", srv.Addr)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("Error starting server: %v\n", err)
			stop()
		}
	}()

	<-ctx.Done()
	fmt.Println("Shutting down gracefully...")
	return srv.Shutdown(context.Background())
}
