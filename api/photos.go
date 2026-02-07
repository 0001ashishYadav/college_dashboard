package api

import (
	"dashboard/db/pgdb"
	"dashboard/token"
	"dashboard/utils"

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
		return InternalServerError("failed to open image")
	}
	defer file.Close()

	// 📝 ALT TEXT
	altTextStr := c.FormValue("alt_text")

	// ☁️ CLOUDINARY UPLOAD (STREAM)
	imageURL, publicID, err := utils.UploadImageStream(
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

	// 💾 SAVE TO DB
	photo, err := server.store.CreatePhoto(
		c.Context(),
		pgdb.CreatePhotoParams{
			ImageUrl:    imageURL,
			AltText:     altText,
			UploadedBy:  int32(payload.ID),
			InstituteID: payload.InstituteID,
			CloudinaryPublicID: pgtype.Text{
				String: publicID,
				Valid:  true,
			},
		},
	)
	if err != nil {
		return InternalServerError(err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(photo)
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

func (server *Server) getPhotosByInstitute(c *fiber.Ctx) error {

	// 1️⃣ Get token payload
	payload, ok := c.Locals(TokenPayloadKey).(*token.TokenPayload)
	if !ok {
		return fiber.NewError(
			fiber.StatusUnauthorized,
			"invalid auth context",
		)
	}

	// 2️⃣ Fetch photos (INSTITUTE SCOPED)
	photos, err := server.store.GetPhotosByInstitute(
		c.Context(),
		payload.InstituteID,
	)
	if err != nil {
		return InternalServerError(err.Error())
	}

	// 3️⃣ Build safe response
	response := make([]fiber.Map, 0, len(photos))

	for _, photo := range photos {

		altText := ""
		if photo.AltText.Valid {
			altText = photo.AltText.String
		}

		response = append(response, fiber.Map{
			"id":           photo.ID,
			"image_url":    photo.ImageUrl,
			"alt_text":     altText,
			"uploaded_by":  photo.UploadedBy,
			"institute_id": photo.InstituteID,
			"created_at":   photo.CreatedAt,
		})
	}

	// 4️⃣ Return response
	return c.JSON(response)
}

func (server *Server) replacePhoto(c *fiber.Ctx) error {

	photoID, err := c.ParamsInt("id")
	if err != nil || photoID <= 0 {
		return fiber.NewError(400, "invalid photo id")
	}

	payload, ok := c.Locals(TokenPayloadKey).(*token.TokenPayload)
	if !ok {
		return fiber.NewError(401, "unauthorized")
	}

	// 1️⃣ Fetch old photo
	oldPhoto, err := server.store.GetPhotoByID(
		c.Context(),
		pgdb.GetPhotoByIDParams{
			ID:          int32(photoID),
			InstituteID: payload.InstituteID,
		},
	)
	if err != nil {
		return fiber.NewError(404, "photo not found")
	}

	// 2️⃣ Get new image
	fileHeader, err := c.FormFile("image")
	if err != nil {
		return fiber.NewError(400, "image file required")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return InternalServerError("failed to open image")
	}
	defer file.Close()

	// 3️⃣ Upload new image
	imageURL, publicID, err := utils.UploadImageStream(
		c.Context(),
		file,
		"institutes/photos",
	)
	if err != nil {
		return InternalServerError("cloudinary upload failed")
	}

	// 4️⃣ Delete old image
	if oldPhoto.CloudinaryPublicID.Valid {
		_ = utils.DeleteImage(
			c.Context(),
			oldPhoto.CloudinaryPublicID.String,
		)
	}

	// 5️⃣ Update DB
	photo, err := server.store.UpdatePhotoImage(
		c.Context(),
		pgdb.UpdatePhotoImageParams{
			ID:          int32(photoID),
			InstituteID: payload.InstituteID,
			ImageUrl:    imageURL,
			CloudinaryPublicID: pgtype.Text{
				String: publicID,
				Valid:  true,
			},
		},
	)
	if err != nil {
		return InternalServerError(err.Error())
	}

	return c.JSON(photo)
}
