package http

import (
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	"github.com/lapitskyss/go_postgresql/internal/video/service"
)

type VideoHandler struct {
	vs     *service.VideoService
	logger *zap.Logger
	tracer opentracing.Tracer
}

func NewVideoHandler(vs *service.VideoService, logger *zap.Logger, tracer opentracing.Tracer) *VideoHandler {
	return &VideoHandler{
		vs:     vs,
		logger: logger,
		tracer: tracer,
	}
}
