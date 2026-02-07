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
