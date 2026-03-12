package handler

import (
	"log/slog"

	"zephyr-storage/internal/service"
)

type Handler struct {
	svc *service.Service
	log *slog.Logger
}

func NewHandler(svc *service.Service, log *slog.Logger) *Handler {
	return &Handler{
		svc: svc,
		log: log,
	}
}
