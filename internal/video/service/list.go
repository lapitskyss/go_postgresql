package service

import (
	"context"
	"fmt"

	"github.com/lapitskyss/go_postgresql/internal/video/storage"
	"github.com/opentracing/opentracing-go"
)

var (
	ErrIncorrectVideoTitle = fmt.Errorf("got an incorrect video title")
	ErrDBRequestFailed     = fmt.Errorf("a request to DB failed")
)

func (s *VideoService) FindVideos(ctx context.Context, title string) ([]*storage.Video, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, s.tracer, "VideoService.Service.FindVideos")
	defer span.Finish()

	if len(title) == 0 {
		return nil, ErrIncorrectVideoTitle
	}

	if len(title) > 255 {
		str := make([]*storage.Video, 0)

		return str, ErrIncorrectVideoTitle
	}

	videos, err := s.db.FindVideosByTitle(ctx, title)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to find video by title: %v", ErrDBRequestFailed, err)
	}

	return videos, nil
}
