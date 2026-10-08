package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

const logFile = ".logs/app.log"

// Init пишет JSON-логи в stdout и, если получится открыть файл, ещё и в .logs/app.log.
func Init(level string) {
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = slog.LevelInfo
	}

	var w io.Writer = os.Stdout
	if err := os.MkdirAll(filepath.Dir(logFile), 0o755); err == nil {
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err == nil {
			w = io.MultiWriter(os.Stdout, f)
		}
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{Level: lvl})))
}
