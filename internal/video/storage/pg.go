package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/opentracing/opentracing-go"
)

type Storage struct {
	db     *pgxpool.Pool
	tracer opentracing.Tracer
}

func NewVideoStorage(db *pgxpool.Pool, tracer opentracing.Tracer) VideoStorage {
	return &Storage{
		db:     db,
		tracer: tracer,
	}
}

func (s *Storage) FindVideosByTitle(ctx context.Context, title string) ([]*Video, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, s.tracer, "VideoService.Storage.FindVideosByTitle")
	defer span.Finish()

	rows, err := s.db.Query(
		ctx,
		`SELECT id, title, description, number_of_views, number_of_likes, created_at FROM videos WHERE title ILIKE $1`,
		"%"+title+"%",
	)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	videos := make([]*Video, 0)
	for rows.Next() {
		v := &Video{}
		if err = rows.Scan(&v.ID, &v.Title, &v.Description, &v.NumberOfViews, &v.NumberOfLikes, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan a received phone: %w", err)
		}
		videos = append(videos, v)
	}

	return videos, nil
}
