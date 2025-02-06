package handlers

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
	"github.com/sirrobot01/lamba/internal/executor"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type Handler struct {
	ex   *executor.Executor
	port string
}

func New(ex *executor.Executor, port string) *Handler {
	return &Handler{
		ex:   ex,
		port: port,
	}
}

func (h *Handler) Start() error {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	srv := &http.Server{
		Addr:    ":" + h.port,
		Handler: r,
	}
	h.Routes(r)

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

func (h *Handler) Routes(r chi.Router) http.Handler {
	r.Group(func(r chi.Router) {
		r.Get("/", h.handleHome)
		r.Post("/invoke", h.Invoker)
		r.Route("/functions", func(r chi.Router) {
			r.Get("/", h.handleFunctionsList)
			r.Post("/", h.handleFunctionsCreate)
			r.Delete("/{id}", h.handleFunctionsDelete)
		})

		r.Route("/events", func(r chi.Router) {
			r.Get("/", h.handleEventList)
			r.Post("/", h.handleEventCreate)
			r.Delete("/{id}", h.handleEventDelete)
		})
	})
	return r
}
