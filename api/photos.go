package api

import (
	"dashboard/db/pgdb"
	"dashboard/token"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgtype"
)

func (server *Server) createPhoto(c *fiber.Ctx) error {
	// 🔐 AUTH
	payload, ok := c.Locals(TokenPayloadKey).(*token.TokenPayload)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	// 📤 FILE
	fileHeader, err := c.FormFile("image")
	if err != nil {
		return fiber.NewError(
			fiber.StatusBadRequest,
			"image file is required",
		)
	}

	file, err := fileHeader.Open()
	if err != nil {
		return InternalServerError("failed to open file")
	}
	defer file.Close()

	// 📝 ALT TEXT
	altTextStr := c.FormValue("alt_text")

	// ☁️ CLOUDINARY UPLOAD
	imageURL, err := server.cloudinary.UploadImage(
		c.Context(),
		file,
		"institutes/photos",
	)
	if err != nil {
		return InternalServerError("cloudinary upload failed")
	}

	// 🧠 Convert alt_text
	altText := pgtype.Text{
		String: altTextStr,
		Valid:  altTextStr != "",
	}

	// 💾 DB SAVE
	photo, err := server.store.CreatePhoto(
		c.Context(),
		pgdb.CreatePhotoParams{
			ImageUrl:    imageURL,
			AltText:     altText,
			UploadedBy:  int32(payload.ID),
			InstituteID: payload.InstituteID,
		},
	)
	if err != nil {
		return InternalServerError(err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":           photo.ID,
		"image_url":    photo.ImageUrl,
		"alt_text":     photo.AltText.String,
		"institute_id": photo.InstituteID,
		"uploaded_by":  photo.UploadedBy,
		"created_at":   photo.CreatedAt,
	})
}

func (server *Server) getPhotoByID(c *fiber.Ctx) error {

	// 1️⃣ Parse photo ID from URL
	photoID, err := c.ParamsInt("id")
	if err != nil || photoID <= 0 {
		return fiber.NewError(
			fiber.StatusBadRequest,
			"invalid photo id",
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

	// 3️⃣ Fetch photo (INSTITUTE SCOPED)
	photo, err := server.store.GetPhotoByID(
		c.Context(),
		pgdb.GetPhotoByIDParams{
			ID:          int32(photoID),
			InstituteID: payload.InstituteID,
		},
	)
	if err != nil {
		if pgdb.ErrorCode(err) == pgdb.ErrorNoRow {
			return NotFoundError("photo not found")
		}
		return InternalServerError(err.Error())
	}

	// 4️⃣ Safe alt_text
	altText := ""
	if photo.AltText.Valid {
		altText = photo.AltText.String
	}

	// 5️⃣ Response
	return c.JSON(fiber.Map{
		"id":           photo.ID,
		"image_url":    photo.ImageUrl,
		"alt_text":     altText,
		"uploaded_by":  photo.UploadedBy,
		"institute_id": photo.InstituteID,
		"created_at":   photo.CreatedAt,
	})
}
