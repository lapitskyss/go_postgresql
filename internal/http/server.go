package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	"github.com/lapitskyss/go_postgresql/internal/store"
	vh "github.com/lapitskyss/go_postgresql/internal/video/http"
	vsrv "github.com/lapitskyss/go_postgresql/internal/video/service"
	vs "github.com/lapitskyss/go_postgresql/internal/video/storage"
)

type Server struct {
	server http.Server
	logger *zap.Logger
	errors chan error
}

func NewServer(db *store.Store, logger *zap.Logger, tracer opentracing.Tracer) *Server {
	r := chi.NewRouter()

	videoStorage := vs.NewVideoStorage(db.Connection(), tracer)
	videoService := vsrv.NewVideoService(videoStorage, tracer)
	videoHTTPHandler := vh.NewVideoHandler(videoService, logger, tracer)

	r.Get("/api/videos/search/{title}", videoHTTPHandler.FindVideos)

	return &Server{
		logger: logger,
		server: http.Server{
			Addr:    ":3000",
			Handler: r,

			ReadTimeout:       1 * time.Second,
			WriteTimeout:      90 * time.Second,
			IdleTimeout:       30 * time.Second,
			ReadHeaderTimeout: 2 * time.Second,
		},
	}
}

func (srv *Server) Start() {
	srv.logger.Info("http server starter")
	go func() {
		err := srv.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			srv.errors <- err
			close(srv.errors)
		}
	}()
}

func (srv *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return srv.server.Shutdown(ctx)
}

func (srv *Server) Notify() <-chan error {
	return srv.errors
}
