package api

import (
	"dashboard/db/pgdb"
	"dashboard/token"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgtype"
)

type CarouselRequest struct {
	Title    string `json:"title" validate:"required"`
	IsActive *bool  `json:"is_active"`
}

func (server *Server) createCarousel(c *fiber.Ctx) error {
	// 🔐 AUTH
	payload, ok := c.Locals(TokenPayloadKey).(*token.TokenPayload)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	// 📥 BODY
	var req CarouselRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if validationErrors := server.validate(req); validationErrors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validationErrors)
	}

	// 🔁 CONVERSIONS
	title := pgtype.Text{
		String: req.Title,
		Valid:  req.Title != "",
	}

	isActive := pgtype.Bool{Valid: false}
	if req.IsActive != nil {
		isActive = pgtype.Bool{
			Bool:  *req.IsActive,
			Valid: true,
		}
	}

	// 💾 DB
	carousel, err := server.store.CreateCarousel(
		c.Context(),
		pgdb.CreateCarouselParams{
			InstituteID: payload.InstituteID,
			Title:       title,
			IsActive:    isActive,
		},
	)
	if err != nil {
		return InternalServerError(err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(carousel)
}

func (server *Server) getCarouselByID(c *fiber.Ctx) error {
	carouselID, err := c.ParamsInt("id")
	if err != nil || carouselID <= 0 {
		return fiber.NewError(400, "invalid carousel id")
	}

	payload := c.Locals(TokenPayloadKey).(*token.TokenPayload)

	rows, err := server.store.GetCarouselWithPhotos(
		c.Context(),
		pgdb.GetCarouselWithPhotosParams{
			ID:          int32(carouselID),
			InstituteID: payload.InstituteID,
		},
	)
	if err != nil || len(rows) == 0 {
		return fiber.NewError(404, "carousel not found")
	}

	// 🧠 Build response
	carousel := fiber.Map{
		"id":           rows[0].CarouselID,
		"institute_id": rows[0].InstituteID,
		"title":        rows[0].Title,
		"is_active":    rows[0].IsActive.Bool,
		"created_at":   rows[0].CreatedAt,
		"photos":       []fiber.Map{},
	}

	photos := []fiber.Map{}

	for _, r := range rows {
		if r.PhotoID.Valid {
			photos = append(photos, fiber.Map{
				"id":            r.PhotoID.Int32,
				"image_url":     r.ImageUrl.String,
				"alt_text":      r.AltText.String,
				"display_text":  r.DisplayText.String,
				"display_order": r.DisplayOrder.Int32,
			})
		}
	}

	carousel["photos"] = photos

	return c.JSON(carousel)
}
