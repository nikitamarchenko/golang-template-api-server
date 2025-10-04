// Package server contain http server related code
package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"

	"github.com/justinas/alice"
)

// Config for HTTP server.
type Config struct {
	Port                            int
	HTTPReadinessProbePeriodSeconds int
	LogLevel                        slog.Level
}

type server struct {
	config         Config
	isShuttingDown atomic.Bool
	rootLogger     *slog.Logger
}

func (s *server) routes() http.Handler {
	return alice.New(
		s.middlewareRecoverPanic,
		s.logRequest,
		commonHeaders,
	).Then(s.handlers())
}

func (s *server) handlers() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealthz())

	return mux
}

// middlewareRecoverPanic intercept app panic.
func (s *server) middlewareRecoverPanic(next http.Handler) http.Handler {
	log := s.newLogger("server.middlewareRecoverPanic")

	getError := func(a any) string {
		switch value := a.(type) {
		case string:
			return value
		case error:
			return value.Error()
		default:
			return fmt.Sprint(value)
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Error("panic", slog.String("err.message", getError(err)))
				w.Header().Set("Connection", "close")
				w.WriteHeader(http.StatusInternalServerError)

				_, err = fmt.Fprintf(
					w,
					`{ "details": "%s" }`,
					http.StatusText(http.StatusInternalServerError),
				)
				if err != nil {
					log.Error("write response", slog.Any("err", err))
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func commonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "Go")
		next.ServeHTTP(w, r)
	})
}

func (s *server) handleHealthz() http.HandlerFunc {
	log := s.newLogger("server.handleHealthz")

	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if s.isShuttingDown.Load() {
			http.Error(w, "shutting down", http.StatusServiceUnavailable)

			return
		}

		_, err := fmt.Fprintln(w, http.StatusText(http.StatusOK))
		if err != nil {
			log.Error("write ok status", slog.Any("err", err))
		}
	})
}

func (s *server) logRequest(next http.Handler) http.Handler {
	log := s.newLogger("server.logRequest")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("request", slog.Any("request", r))
		next.ServeHTTP(w, r)
	})
}
