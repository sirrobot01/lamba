package invoker

import (
	"errors"
	"fmt"
	"github.com/sirrobot01/lamba/pkg/executor"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type HTTPInvoker struct {
	port   string
	ex     *executor.Executor
	server *http.Server
	logger *log.Logger
}

// NewHTTPInvoker creates a new Invoker instance
func NewHTTPInvoker(ex *executor.Executor, port string) Invoker {
	return &HTTPInvoker{
		ex:     ex,
		port:   port,
		logger: log.New(os.Stdout, "HTTP: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Invoke executes a function with the given name and payload

func (h *HTTPInvoker) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/invoke", h.loggingMiddleware(func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("function")
		body, _ := io.ReadAll(r.Body)
		result, err := h.ex.Execute("http", name, body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(result)
	}))

	h.server = &http.Server{
		Addr:    ":" + h.port,
		Handler: mux,
	}

	h.logger.Println("Starting HTTP server on :%s", h.port)
	if err := h.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("HTTP server error: %v", err)
	}
	return nil
}

func (h *HTTPInvoker) loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom ResponseWriter to capture the status code
		rw := &responseWriter{w, http.StatusOK}

		next.ServeHTTP(rw, r)

		h.logger.Printf(
			"%s %s %s %d %v",
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			rw.statusCode,
			time.Since(start),
		)
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (h *HTTPInvoker) Stop() error {
	h.logger.Println("Stopping HTTP server")
	return h.server.Shutdown(nil)
}
