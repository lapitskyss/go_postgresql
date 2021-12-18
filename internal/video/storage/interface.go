package storage

import (
	"context"
	"time"
)

type VideoStorage interface {
	FindVideosByTitle(ctx context.Context, title string) ([]*Video, error)
}

type Video struct {
	ID            int64
	Title         string
	Description   string
	NumberOfViews int64
	NumberOfLikes int64
	CreatedAt     time.Time
}
