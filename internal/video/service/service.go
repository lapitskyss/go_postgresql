package service

import (
	"github.com/opentracing/opentracing-go"

	"github.com/lapitskyss/go_postgresql/internal/video/storage"
)

type VideoService struct {
	db     storage.VideoStorage
	tracer opentracing.Tracer
}

func NewVideoService(db storage.VideoStorage, tracer opentracing.Tracer) *VideoService {
	return &VideoService{
		db:     db,
		tracer: tracer,
	}
}
