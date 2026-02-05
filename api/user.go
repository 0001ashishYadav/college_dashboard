package api

import (
	"dashboard/db/pgdb"
	"dashboard/token"
	"time"

	"github.com/gofiber/fiber/v2"
)

type userLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userLoginResponse struct {
	Token          string    `json:"token"`
	TokenExpiresAt time.Time `json:"token_expires_at"`
	ID             int64     `json:"id"`
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	Role           string    `json:"role"`
	InstituteID    int32     `json:"institute_id"`
}

func (server *Server) userLogin(c *fiber.Ctx) error {
	var req userLoginRequest

	if err := c.BodyParser(&req); err != nil {
		return err
	}

	if validationErrors := server.validate(req); validationErrors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validationErrors)
	}

	user, err := server.store.LoginUser(
		c.Context(),
		pgdb.LoginUserParams{
			Email: req.Email,
		},
	)
	if err != nil {
		if pgdb.ErrorCode(err) == pgdb.ErrorNoRow {
			return NotFoundError("invalid email or password")
		}
		return InternalServerError(err.Error())
	}

	// ⚠️ You should hash later — keeping as-is for now
	if user.Password != req.Password {
		return fiber.NewError(
			fiber.StatusUnauthorized,
			"invalid email or password",
		)
	}

	// 🔥 CREATE TOKEN WITH institute_id
	token, payload, err := server.token.CreateToken(
		int64(user.ID),
		user.Email,
		user.Role.String,
		user.Name,
		user.InstituteID,
		server.config.TokenDuration,
	)
	if err != nil {
		return InternalServerError("failed to generate token")
	}

	return c.JSON(userLoginResponse{
		Token:          token,
		TokenExpiresAt: payload.ExpiredAt,
		ID:             payload.ID,
		Email:          payload.Email,
		Role:           payload.Role,
		Name:           payload.Name,
		InstituteID:    payload.InstituteID,
	})
}

func (server *Server) getUserByID(c *fiber.Ctx) error {

	// 1️⃣ Read param
	userID, err := c.ParamsInt("id")
	if err != nil || userID <= 0 {
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: "invalid user id",
		}
	}

	// 2️⃣ Get token payload safely
	payload, ok := c.Locals(TokenPayloadKey).(*token.TokenPayload)
	if !ok {
		return &fiber.Error{
			Code:    fiber.StatusUnauthorized,
			Message: "invalid auth context",
		}
	}

	// 3️⃣ Fetch user
	user, err := server.store.GetUserByID(
		c.Context(),
		pgdb.GetUserByIDParams{
			ID:          int32(userID),
			InstituteID: payload.InstituteID, // ⚠️ use correct field
		},
	)

	if err != nil {
		if pgdb.ErrorCode(err) == pgdb.ErrorNoRow {
			return NotFoundError("user not found")
		}
		return InternalServerError(err.Error())
	}

	// 4️⃣ Safe role
	role := ""
	if user.Role.Valid {
		role = user.Role.String
	}

	// 5️⃣ Response
	return c.JSON(fiber.Map{
		"id":           user.ID,
		"institute_id": user.InstituteID,
		"name":         user.Name,
		"email":        user.Email,
		"role":         role,
		"is_active":    user.IsActive,
		"created_at":   user.CreatedAt,
	})
}

func (server *Server) getUserByEmail(c *fiber.Ctx) error {
	// 1️⃣ Read email from query param
	email := c.Query("email")

	if email == "" {
		return &fiber.Error{
			Code:    fiber.StatusBadRequest,
			Message: "email is required",
		}
	}

	// 2️⃣ Fetch user from DB
	user, err := server.store.GetUserByEmail(
		c.Context(),
		email,
	)

	if err != nil {
		if pgdb.ErrorCode(err) == pgdb.ErrorNoRow {
			return NotFoundError("user not found")
		}
		return InternalServerError(err.Error())
	}

	// 3️⃣ Return safe response (NO password)
	return c.JSON(fiber.Map{
		"id":           user.ID,
		"institute_id": user.InstituteID,
		"name":         user.Name,
		"email":        user.Email,
		"role":         user.Role.String,
		"is_active":    user.IsActive,
		"created_at":   user.CreatedAt,
	})
}
