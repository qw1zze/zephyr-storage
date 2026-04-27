package handler

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"zephyr-storage/internal/service"
)

type blobService interface {
	Upload(ctx context.Context, data []byte) (string, error)
	Get(ctx context.Context, cid string) ([]byte, error)
}

type profileService interface {
	SaveProfile(ctx context.Context, name, avatar string) (string, error)
	GetProfile(ctx context.Context, cid string) (*service.Profile, error)
}

type Handler struct {
	svc        blobService
	profileSvc profileService
	maxSizeB   int64
	log        *slog.Logger
}

func NewHandler(svc blobService, profileSvc profileService, maxSizeMB int64, log *slog.Logger) *Handler {
	return &Handler{
		svc:        svc,
		profileSvc: profileSvc,
		maxSizeB:   maxSizeMB * 1024 * 1024,
		log:        log,
	}
}

func errorResponse(c *fiber.Ctx, status int, msg string) error {
	return c.Status(status).JSON(fiber.Map{"error": msg})
}
