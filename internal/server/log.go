package server

import (
	"log/slog"

	slogformatter "github.com/samber/slog-formatter"
	slogmulti "github.com/samber/slog-multi"
)

func setupLogger(log *slog.Logger) *slog.Logger {
	return slog.New(
		slogmulti.
			Pipe(slogformatter.NewFormatterHandler(
				slogformatter.ErrorFormatter("err"),
				slogformatter.HTTPRequestFormatter(false),
			)).
			Handler(log.Handler()),
	)
}

func (s *server) newLogger(funcName string) *slog.Logger {
	return s.rootLogger.With("func", funcName)
}
