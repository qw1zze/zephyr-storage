package handler

import (
	"bytes"
	"errors"
	"io"

	"github.com/gofiber/fiber/v2"

	"zephyr-storage/internal/ipfs"
	"zephyr-storage/internal/service"
)

func (h *Handler) UploadBlob(c *fiber.Ctx) error {
	if c.Get("Content-Type") != "application/octet-stream" {
		return errorResponse(c, fiber.StatusUnsupportedMediaType,
			"content-type must be application/octet-stream")
	}

	limit := h.maxSizeB + 1
	lr := io.LimitReader(bytes.NewReader(c.Body()), limit)

	data, err := io.ReadAll(lr)
	if err != nil {
		h.log.Error("handler: upload: read body", "error", err)
		return errorResponse(c, fiber.StatusInternalServerError, "internal server error")
	}

	if int64(len(data)) == limit {
		return errorResponse(c, fiber.StatusRequestEntityTooLarge, "payload too large")
	}

	cid, err := h.svc.Upload(c.UserContext(), data)
	if err != nil {
		h.log.Error("handler: upload: service", "error", err)
		if errors.Is(err, service.ErrTooLarge) || errors.Is(err, ipfs.ErrTooLarge) {
			return errorResponse(c, fiber.StatusRequestEntityTooLarge, "payload too large")
		}
		return errorResponse(c, fiber.StatusInternalServerError, "internal server error")
	}

	return c.JSON(fiber.Map{"cid": cid})
}

func (h *Handler) GetBlob(c *fiber.Ctx) error {
	cid := c.Params("cid")

	data, err := h.svc.Get(c.UserContext(), cid)
	if err != nil {
		h.log.Error("handler: get: service", "error", err, "cid", cid)

		switch {
		case errors.Is(err, service.ErrInvalidCID), errors.Is(err, service.ErrEmptyCID):
			return errorResponse(c, fiber.StatusBadRequest, "invalid cid format")
		case errors.Is(err, ipfs.ErrNotFound):
			return errorResponse(c, fiber.StatusNotFound, "blob not found")
		case errors.Is(err, service.ErrTooLarge), errors.Is(err, ipfs.ErrTooLarge):
			return errorResponse(c, fiber.StatusRequestEntityTooLarge, "payload too large")
		default:
			return errorResponse(c, fiber.StatusInternalServerError, "internal server error")
		}
	}

	c.Set("Content-Type", "application/octet-stream")
	return c.Send(data)
}
