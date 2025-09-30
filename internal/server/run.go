package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os/signal"
	"os/user"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	httpReadHeaderTimeout  = 15 * time.Second
	httpShutdownPeriod     = 15 * time.Second
	httpShutdownHardPeriod = 3 * time.Second
	/*
		readinessProbe:
			httpGet:
				path: /healthz
				port: 8482
			periodSeconds: 1
		delay = httpReadinessDrainDelay + httpReadinessProbePeriodSeconds
	*/
	httpReadinessDrainDelay = 1 * time.Second
)

func newServer(logger *slog.Logger, config Config) server {
	return server{
		config:         config,
		rootLogger:     setupLogger(logger),
		isShuttingDown: atomic.Bool{},
	}
}

// Run start HTTP server with provided config.
func Run( //nolint:funlen // too many closures
	logger *slog.Logger,
	config Config,
	allowRootUser bool,
) error {
	s := newServer(logger, config)
	log := s.newLogger("server.Run")

	err := checkUser()
	if err != nil {
		if !allowRootUser {
			return err
		}

		log.Warn("server run as root")
	}

	log.Info("init",
		slog.Group("config",
			slog.Group("http",
				slog.Int("port", config.Port),
				slog.Int("readinessProbe.periodSeconds", config.HTTPReadinessProbePeriodSeconds),
			),
		),
		slog.Group("log", slog.Any("level", config.LogLevel)),
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// global context for all connections for shutdown propagation
	ongoingCtx, stopOngoingGracefully := context.WithCancel(context.Background())
	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: s.routes(),
		BaseContext: func(_ net.Listener) context.Context {
			return ongoingCtx
		},
		ReadHeaderTimeout: httpReadHeaderTimeout,
	}
	listenAndServeFailed := make(chan error)

	go func() {
		defer close(listenAndServeFailed)

		log.Info("HTTP server run")

		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			listenAndServeFailed <- err
		}
	}()

	select {
	case <-ctx.Done(): // wait for sigint or sigterm:
		break
	case err := <-listenAndServeFailed: // in case of srv.ListenAndServe failed
		log.Error("HTTP server", slog.Any("err", err))
		stopOngoingGracefully()

		return err
	}

	log.Info("shutdown initiated")
	stop()
	s.isShuttingDown.Store(true) // fail readinessProbe

	delay := httpReadinessDrainDelay + time.Duration(
		config.HTTPReadinessProbePeriodSeconds,
	)*time.Second
	log.Info(
		"wait for readinessProbe mark pod as NotReady",
		slog.Any("delay.seconds", delay/time.Second),
	)
	time.Sleep(delay)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), httpShutdownPeriod)
	defer cancel()

	err = srv.Shutdown(shutdownCtx)

	stopOngoingGracefully()

	if err != nil {
		log.Error("server shutdown failed", slog.Any("err", err))
		time.Sleep(httpShutdownHardPeriod)

		return fmt.Errorf("server shutdown: %w", err)
	}

	log.Info("quit")

	return nil
}

// ErrRootUser caused container run as root user.
var ErrRootUser = errors.New("server runs as root")

func checkUser() error {
	u, err := user.Current()
	if err != nil {
		return fmt.Errorf("get user: %w", err)
	}

	if u.Username == "root" {
		return ErrRootUser
	}

	if u.Uid == "0" {
		return fmt.Errorf("%w: UID(0)", ErrRootUser)
	}

	if u.Gid == "0" {
		return fmt.Errorf("%w: GID(0)", ErrRootUser)
	}

	return nil
}
