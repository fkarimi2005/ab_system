package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func New(port string, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:              ":" + port,
			Handler:           handler,
			MaxHeaderBytes:    1 << 20, // 1 МБ
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      10 * time.Second,
		},
	}
}

func (s *Server) Run() error {
	slog.Info("http server started", "addr", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}
