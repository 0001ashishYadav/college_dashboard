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
