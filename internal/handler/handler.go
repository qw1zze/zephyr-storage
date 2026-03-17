package handler

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

type blobService interface {
	Upload(ctx context.Context, data []byte) (string, error)
	Get(ctx context.Context, cid string) ([]byte, error)
}

type Handler struct {
	svc      blobService
	maxSizeB int64
	log      *slog.Logger
}

func NewHandler(svc blobService, maxSizeMB int64, log *slog.Logger) *Handler {
	return &Handler{
		svc:      svc,
		maxSizeB: maxSizeMB * 1024 * 1024,
		log:      log,
	}
}

func errorResponse(c *fiber.Ctx, status int, msg string) error {
	return c.Status(status).JSON(fiber.Map{"error": msg})
}
