package api

import (
	"dashboard/db/pgdb"
	"dashboard/token"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateNoticeRequest struct {
	Title       string     `json:"title" validate:"required,min=3"`
	Description string     `json:"description"`
	IsPublished bool       `json:"is_published"`
	PublishDate *time.Time `json:"publish_date"`
}

func (server *Server) createNotice(c *fiber.Ctx) error {

	// 1️⃣ Parse request body
	var req CreateNoticeRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(
			fiber.StatusBadRequest,
			"invalid request body",
		)
	}

	// 2️⃣ Validate request
	if validationErrors := server.validate(req); validationErrors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validationErrors)
	}

	// 3️⃣ Get token payload
	payload, ok := c.Locals(TokenPayloadKey).(*token.TokenPayload)
	if !ok {
		return fiber.NewError(
			fiber.StatusUnauthorized,
			"invalid auth context",
		)
	}

	// 4️⃣ Admin-only access
	if payload.Role != "admin" {
		return fiber.NewError(
			fiber.StatusForbidden,
			"admin access required",
		)
	}

	// 5️⃣ Create notice
	notice, err := server.store.CreateNotice(
		c.Context(),
		pgdb.CreateNoticeParams{
			InstituteID: payload.InstituteID,
			Title:       req.Title,
			Description: pgtype.Text{String: req.Description, Valid: req.Description != ""},
			IsPublished: pgtype.Bool{Bool: req.IsPublished, Valid: true},
			PublishDate: pgtype.Date{
				Time: func() time.Time {
					if req.PublishDate != nil {
						return *req.PublishDate
					}
					return time.Time{}
				}(),
				Valid: req.PublishDate != nil,
			},
		},
	)
	if err != nil {
		return InternalServerError(err.Error())
	}

	// 6️⃣ Response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":           notice.ID,
		"institute_id": notice.InstituteID,
		"title":        notice.Title,
		"description":  notice.Description,
		"is_published": notice.IsPublished,
		"publish_date": notice.PublishDate,
		"created_at":   notice.CreatedAt,
	})
}
