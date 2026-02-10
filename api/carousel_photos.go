package api

import (
	"dashboard/db/pgdb"
	"dashboard/token"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgtype"
)

// 📥 Create carousel photo
type CreateCarouselPhotoRequest struct {
	PhotoID      int32  `json:"photo_id" validate:"required"`
	DisplayText  string `json:"display_text"`
	DisplayOrder int32  `json:"display_order"`
}

// 📥 Update carousel photo
type UpdateCarouselPhotoRequest struct {
	DisplayText  string `json:"display_text"`
	DisplayOrder int32  `json:"display_order" validate:"required"`
}

// 📥 Reorder carousel photo
type ReorderCarouselPhotoRequest struct {
	DisplayOrder int32 `json:"display_order" validate:"required"`
}

func (server *Server) createCarouselPhoto(c *fiber.Ctx) error {

	// 🔐 AUTH
	payload, ok := c.Locals(TokenPayloadKey).(*token.TokenPayload)
	if !ok {
		return fiber.NewError(401, "unauthorized")
	}

	// 📌 Carousel ID from URL
	carouselID, err := c.ParamsInt("id")
	if err != nil || carouselID <= 0 {
		return fiber.NewError(400, "invalid carousel id")
	}

	// 📥 BODY
	var req CreateCarouselPhotoRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(400, "invalid request body")
	}

	if errs := server.validate(req); errs != nil {
		return c.Status(400).JSON(errs)
	}

	// 🔒 INSTITUTE OWNERSHIP CHECK (CRITICAL)
	rows, err := server.store.GetCarouselWithPhotos(
		c.Context(),
		pgdb.GetCarouselWithPhotosParams{
			ID:          int32(carouselID),
			InstituteID: payload.InstituteID,
		},
	)
	if err != nil || len(rows) == 0 {
		return fiber.NewError(
			fiber.StatusForbidden,
			"carousel not found or access denied",
		)
	}

	// 🧠 Convert types
	displayText := pgtype.Text{
		String: req.DisplayText,
		Valid:  req.DisplayText != "",
	}

	displayOrder := pgtype.Int4{
		Int32: req.DisplayOrder,
		Valid: true,
	}

	// 💾 INSERT
	photo, err := server.store.CreateCarouselPhoto(
		c.Context(),
		pgdb.CreateCarouselPhotoParams{
			CarouselID:   int32(carouselID),
			PhotoID:      req.PhotoID,
			DisplayText:  displayText,
			DisplayOrder: displayOrder,
		},
	)
	if err != nil {
		return InternalServerError(err.Error())
	}

	return c.Status(201).JSON(photo)
}

func (server *Server) getCarouselPhotoByID(c *fiber.Ctx) error {
	payload := c.Locals(TokenPayloadKey).(*token.TokenPayload)

	id, err := c.ParamsInt("id")
	if err != nil || id <= 0 {
		return fiber.NewError(400, "invalid carousel photo id")
	}

	row, err := server.store.GetCarouselPhotoWithImage(
		c.Context(),
		int32(id),
	)
	if err != nil {
		return NotFoundError("carousel photo not found")
	}

	// 🔐 Institute security
	_, err = server.store.GetCarouselWithPhotos(
		c.Context(),
		pgdb.GetCarouselWithPhotosParams{
			ID:          row.CarouselID,
			InstituteID: payload.InstituteID,
		},
	)
	if err != nil {
		return fiber.NewError(403, "access denied")
	}

	return c.JSON(fiber.Map{
		"id":            row.ID,
		"carousel_id":   row.CarouselID,
		"photo_id":      row.PhotoID,
		"display_text":  row.DisplayText.String,
		"display_order": row.DisplayOrder.Int32,
		"image_url":     row.ImageUrl,
		"alt_text":      row.AltText.String,
		"created_at":    row.CreatedAt,
	})
}
