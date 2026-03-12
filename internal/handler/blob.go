package handler

import "github.com/gofiber/fiber/v2"

func (h *Handler) UploadBlob(c *fiber.Ctx) error {
	cid, err := h.svc.Upload(c.UserContext(), c.Body())
	if err != nil {
		h.log.Error("failed to upload blob", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	return c.JSON(fiber.Map{
		"cid": cid,
	})
}

func (h *Handler) GetBlob(c *fiber.Ctx) error {
	cid := c.Params("cid")

	data, err := h.svc.Get(c.UserContext(), cid)
	if err != nil {
		h.log.Error("failed to get blob", "error", err, "cid", cid)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	if data == nil {
		return c.JSON(fiber.Map{
			"data": "not_implemented",
		})
	}

	return c.Send(data)
}
