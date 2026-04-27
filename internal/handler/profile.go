package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"zephyr-storage/internal/ipfs"
	"zephyr-storage/internal/service"
)

func (h *Handler) SaveProfile(c *fiber.Ctx) error {
	var body struct {
		Name   string `json:"name"`
		Avatar string `json:"avatar"`
	}
	if err := c.BodyParser(&body); err != nil {
		return errorResponse(c, fiber.StatusBadRequest, "invalid json body")
	}

	cid, err := h.profileSvc.SaveProfile(c.UserContext(), body.Name, body.Avatar)
	if err != nil {
		h.log.Error("handler: save profile: service", "error", err)
		if errors.Is(err, service.ErrEmptyName) {
			return errorResponse(c, fiber.StatusBadRequest, "name must not be empty")
		}
		return errorResponse(c, fiber.StatusInternalServerError, "internal server error")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"cid": cid})
}

func (h *Handler) GetProfile(c *fiber.Ctx) error {
	cid := c.Params("cid")

	profile, err := h.profileSvc.GetProfile(c.UserContext(), cid)
	if err != nil {
		h.log.Error("handler: get profile: service", "error", err, "cid", cid)
		switch {
		case errors.Is(err, service.ErrInvalidCID), errors.Is(err, service.ErrEmptyCID):
			return errorResponse(c, fiber.StatusBadRequest, "invalid cid format")
		case errors.Is(err, ipfs.ErrNotFound):
			return errorResponse(c, fiber.StatusNotFound, "profile not found")
		default:
			return errorResponse(c, fiber.StatusInternalServerError, "internal server error")
		}
	}

	return c.JSON(profile)
}
