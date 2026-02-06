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

type UpdateNoticeRequest struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
	IsPublished *bool  `json:"is_published"`
	PublishDate string `json:"publish_date"` // YYYY-MM-DD
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

func (server *Server) getNoticeByID(c *fiber.Ctx) error {

	// 1️⃣ Read notice ID
	noticeID, err := c.ParamsInt("id")
	if err != nil || noticeID <= 0 {
		return fiber.NewError(
			fiber.StatusBadRequest,
			"invalid notice id",
		)
	}

	// 2️⃣ Get token payload
	payload, ok := c.Locals(TokenPayloadKey).(*token.TokenPayload)
	if !ok {
		return fiber.NewError(
			fiber.StatusUnauthorized,
			"invalid auth context",
		)
	}

	// 3️⃣ Fetch notice (USE CORRECT sqlc METHOD)
	notice, err := server.store.GetNotice(
		c.Context(),
		pgdb.GetNoticeParams{
			ID:          int32(noticeID),
			InstituteID: payload.InstituteID,
		},
	)
	if err != nil {
		if pgdb.ErrorCode(err) == pgdb.ErrorNoRow {
			return NotFoundError("notice not found")
		}
		return InternalServerError(err.Error())
	}

	// 4️⃣ Response
	return c.JSON(fiber.Map{
		"id":           notice.ID,
		"institute_id": notice.InstituteID,
		"title":        notice.Title,
		"description":  notice.Description,
		"is_published": notice.IsPublished,
		"publish_date": notice.PublishDate,
		"created_at":   notice.CreatedAt,
	})
}

func (server *Server) getNoticesByInstitute(c *fiber.Ctx) error {

	// 1️⃣ Get token payload
	payload, ok := c.Locals(TokenPayloadKey).(*token.TokenPayload)
	if !ok {
		return fiber.NewError(
			fiber.StatusUnauthorized,
			"invalid auth context",
		)
	}

	// 2️⃣ Fetch notices
	notices, err := server.store.GetNoticesByInstitute(
		c.Context(),
		payload.InstituteID,
	)
	if err != nil {
		return InternalServerError(err.Error())
	}

	// 3️⃣ Build response
	response := make([]fiber.Map, 0, len(notices))

	for _, notice := range notices {
		response = append(response, fiber.Map{
			"id":           notice.ID,
			"institute_id": notice.InstituteID,
			"title":        notice.Title,
			"description":  notice.Description,
			"is_published": notice.IsPublished,
			"publish_date": notice.PublishDate,
			"created_at":   notice.CreatedAt,
		})
	}

	// 4️⃣ Return response
	return c.JSON(response)
}

func (server *Server) updateNotice(c *fiber.Ctx) error {

	// 1️⃣ Parse notice ID from URL
	noticeID, err := c.ParamsInt("id")
	if err != nil || noticeID <= 0 {
		return fiber.NewError(
			fiber.StatusBadRequest,
			"invalid notice id",
		)
	}

	// 2️⃣ Parse request body
	var req UpdateNoticeRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(
			fiber.StatusBadRequest,
			"invalid request body",
		)
	}

	// 3️⃣ Validate request
	if validationErrors := server.validate(req); validationErrors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validationErrors)
	}

	// 4️⃣ Get token payload
	payload, ok := c.Locals(TokenPayloadKey).(*token.TokenPayload)
	if !ok {
		return fiber.NewError(
			fiber.StatusUnauthorized,
			"invalid auth context",
		)
	}

	// 🔐 ADMIN CHECK (ADDED)
	if payload.Role != "admin" {
		return fiber.NewError(
			fiber.StatusForbidden,
			"admin access required",
		)
	}

	// 5️⃣ Convert description to pgtype.Text
	desc := pgtype.Text{
		String: req.Description,
		Valid:  req.Description != "",
	}

	// 6️⃣ Convert is_published to pgtype.Bool
	isPublished := pgtype.Bool{Valid: false}
	if req.IsPublished != nil {
		isPublished = pgtype.Bool{
			Bool:  *req.IsPublished,
			Valid: true,
		}
	}

	// 7️⃣ Convert publish_date to pgtype.Date
	publishDate := pgtype.Date{Valid: false}
	if req.PublishDate != "" {
		if err := publishDate.Scan(req.PublishDate); err != nil {
			return fiber.NewError(
				fiber.StatusBadRequest,
				"publish_date must be YYYY-MM-DD",
			)
		}
	}

	// 8️⃣ Update notice
	notice, err := server.store.UpdateNotice(
		c.Context(),
		pgdb.UpdateNoticeParams{
			ID:          int32(noticeID),
			Title:       req.Title,
			Description: desc,
			IsPublished: isPublished,
			PublishDate: publishDate,
		},
	)
	if err != nil {
		if pgdb.ErrorCode(err) == pgdb.ErrorNoRow {
			return NotFoundError("notice not found")
		}
		return InternalServerError(err.Error())
	}

	// 9️⃣ Institute ownership check (SECURITY)
	if notice.InstituteID != payload.InstituteID {
		return fiber.NewError(
			fiber.StatusForbidden,
			"you are not allowed to update this notice",
		)
	}

	// 🔟 Response
	return c.JSON(fiber.Map{
		"id":           notice.ID,
		"institute_id": notice.InstituteID,
		"title":        notice.Title,
		"description":  notice.Description.String,
		"is_published": notice.IsPublished.Bool,
		"publish_date": notice.PublishDate.Time,
		"created_at":   notice.CreatedAt,
	})
}

func (server *Server) deleteNotice(c *fiber.Ctx) error {

	// 1️⃣ Parse notice ID from URL
	noticeID, err := c.ParamsInt("id")
	if err != nil || noticeID <= 0 {
		return fiber.NewError(
			fiber.StatusBadRequest,
			"invalid notice id",
		)
	}

	// 2️⃣ Get token payload
	payload, ok := c.Locals(TokenPayloadKey).(*token.TokenPayload)
	if !ok {
		return fiber.NewError(
			fiber.StatusUnauthorized,
			"invalid auth context",
		)
	}

	// 🔐 3️⃣ Admin-only access
	if payload.Role != "admin" {
		return fiber.NewError(
			fiber.StatusForbidden,
			"admin access required",
		)
	}

	// 4️⃣ Fetch notice first (SECURITY CHECK)
	notice, err := server.store.GetNotice(
		c.Context(),
		pgdb.GetNoticeParams{
			ID:          int32(noticeID),
			InstituteID: payload.InstituteID,
		},
	)
	if err != nil {
		if pgdb.ErrorCode(err) == pgdb.ErrorNoRow {
			return NotFoundError("notice not found")
		}
		return InternalServerError(err.Error())
	}

	// 5️⃣ Delete notice
	if err := server.store.DeleteNotice(
		c.Context(),
		notice.ID,
	); err != nil {
		return InternalServerError(err.Error())
	}

	// 6️⃣ Success response
	return c.JSON(fiber.Map{
		"message":   "notice deleted successfully",
		"notice_id": notice.ID,
	})
}
