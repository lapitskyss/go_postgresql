package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	"github.com/lapitskyss/go_postgresql/internal/video/service"
)

func (h *VideoHandler) FindVideos(w http.ResponseWriter, r *http.Request) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(r.Context(), h.tracer, "VideoService.Handler.FindVideos")
	defer span.Finish()

	title := chi.URLParam(r, "title")
	videos, err := h.vs.FindVideos(ctx, title)
	if err != nil {
		h.logger.Error("Error to call function FindVideos", zap.NamedError("err", err))
		if errors.Is(err, service.ErrIncorrectVideoTitle) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrDBRequestFailed) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(videos)
	if err != nil {
		h.logger.Error("Failed to serialize the video list to JSON", zap.NamedError("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(resp); err != nil {
		h.logger.Error("Failed to write the video list as a response body", zap.NamedError("err", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
