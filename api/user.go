package api

import (
	"dashboard/db/pgdb"
	"dashboard/token"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgtype"
)

type userLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
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

// ✅ Create user request (ADMIN)
type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=3"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Role     string `json:"role" validate:"required"`
	IsActive bool   `json:"is_active"`
}

type UpdateUserRequest struct {
	Name     string `json:"name" validate:"required,min=3"`
	Role     string `json:"role" validate:"required"`
	IsActive bool   `json:"is_active"`
}

type UpdateUserPasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

func (server *Server) createUser(c *fiber.Ctx) error {

	// 1️⃣ Parse request body
	var req CreateUserRequest
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

	// 5️⃣ Create user (in same institute)
	user, err := server.store.CreateUser(
		c.Context(),
		pgdb.CreateUserParams{
			InstituteID: payload.InstituteID,
			Name:        req.Name,
			Email:       req.Email,
			Password:    req.Password, // ⚠️ hash later

			Role: pgtype.Text{
				String: req.Role,
				Valid:  true,
			},
			IsActive: pgtype.Bool{
				Bool:  req.IsActive,
				Valid: true,
			},
		},
	)

	if err != nil {
		// unique email violation
		if pgdb.ErrorCode(err) == pgdb.ErrorDuplicateKey {
			return fiber.NewError(
				fiber.StatusConflict,
				"email already exists",
			)
		}
		return InternalServerError(err.Error())
	}

	// 6️⃣ Safe role
	role := ""
	if user.Role.Valid {
		role = user.Role.String
	}

	// 7️⃣ Response (NO password)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":           user.ID,
		"institute_id": user.InstituteID,
		"name":         user.Name,
		"email":        user.Email,
		"role":         role,
		"is_active":    user.IsActive,
		"created_at":   user.CreatedAt,
	})
}

func (server *Server) updateUser(c *fiber.Ctx) error {
	// 1️⃣ Get user_id from URL
	userID, err := c.ParamsInt("id")
	if err != nil || userID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid user id")
	}

	// 2️⃣ Parse request body
	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	// 3️⃣ Validate request
	if err := server.valid.Struct(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// 4️⃣ Get institute_id from JWT
	authPayload := c.Locals("authPayload").(*token.TokenPayload)

	// 5️⃣ Prepare DB params (sqlc + pgtype)
	arg := pgdb.UpdateUserParams{
		ID:          int32(userID),
		Name:        req.Name,
		Role:        pgtype.Text{String: req.Role, Valid: true},
		IsActive:    pgtype.Bool{Bool: req.IsActive, Valid: true},
		InstituteID: authPayload.InstituteID, // 🔐 institute safety
	}

	// 6️⃣ Execute query
	user, err := server.store.UpdateUser(c.Context(), arg)
	if err != nil {

		switch pgdb.ErrorCode(err) {

		case pgdb.ErrorNoRow:
			return fiber.NewError(
				fiber.StatusNotFound,
				"user not found",
			)

		case pgdb.ErrorDuplicateKey:
			return fiber.NewError(
				fiber.StatusConflict,
				"duplicate value",
			)
		}

		return InternalServerError(err.Error())
	}

	// 7️⃣ Response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "user updated successfully",
		"user": fiber.Map{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"role":       user.Role.String,
			"is_active":  user.IsActive.Bool,
			"institute":  user.InstituteID,
			"updated_at": user.UpdatedAt,
		},
	})
}

func (server *Server) UpdateUserPassword(c *fiber.Ctx) error {

	// 1️⃣ Parse user ID from URL
	userID, err := c.ParamsInt("id")
	if err != nil || userID <= 0 {
		return fiber.NewError(
			fiber.StatusBadRequest,
			"invalid user id",
		)
	}

	// 2️⃣ Parse request body
	var req UpdateUserPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(
			fiber.StatusBadRequest,
			"invalid request body",
		)
	}

	// 3️⃣ Validate request
	if errs := server.validate(req); errs != nil {
		return c.Status(fiber.StatusBadRequest).JSON(errs)
	}

	// 4️⃣ Get JWT payload
	payload, ok := c.Locals(TokenPayloadKey).(*token.TokenPayload)
	if !ok {
		return fiber.NewError(
			fiber.StatusUnauthorized,
			"invalid auth context",
		)
	}

	// 5️⃣ Only admin OR self user can update password
	if payload.Role != "admin" && int64(userID) != payload.ID {
		return fiber.NewError(
			fiber.StatusForbidden,
			"not allowed to update this user's password",
		)
	}

	// 6️⃣ Fetch existing user
	user, err := server.store.GetUserByID(
		c.Context(),
		pgdb.GetUserByIDParams{
			ID:          int32(userID),
			InstituteID: payload.InstituteID,
		},
	)
	if err != nil {
		if pgdb.ErrorCode(err) == pgdb.ErrorNoRow {
			return NotFoundError("user not found")
		}
		return InternalServerError(err.Error())
	}

	// 7️⃣ Verify old password
	// ⚠️ Plain compare for now (hash later)
	if user.Password != req.OldPassword {
		return fiber.NewError(
			fiber.StatusUnauthorized,
			"old password is incorrect",
		)
	}

	// 8️⃣ Update password
	updatedUser, err := server.store.UpdateUserPassword(
		c.Context(),
		pgdb.UpdateUserPasswordParams{
			Password:    req.NewPassword,
			ID:          int32(userID),
			InstituteID: payload.InstituteID,
		},
	)
	if err != nil {
		return InternalServerError(err.Error())
	}

	// 9️⃣ Success response
	return c.JSON(fiber.Map{
		"message":    "password updated successfully",
		"user_id":    updatedUser.ID,
		"updated_at": updatedUser.UpdatedAt,
	})
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

	// 1️⃣ Read email
	email := c.Query("email")
	if email == "" {
		return fiber.NewError(
			fiber.StatusBadRequest,
			"email query parameter is required",
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

	// 3️⃣ Fetch user (INSTITUTE SCOPED)
	user, err := server.store.GetUserByEmail(
		c.Context(),
		pgdb.GetUserByEmailParams{
			Email:       email,
			InstituteID: payload.InstituteID,
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

func (server *Server) getUsersByInstitute(c *fiber.Ctx) error {

	// 1️⃣ Get token payload
	payload, ok := c.Locals(TokenPayloadKey).(*token.TokenPayload)
	if !ok {
		return fiber.NewError(
			fiber.StatusUnauthorized,
			"invalid auth context",
		)
	}

	// 2️⃣ (Optional but recommended) Admin-only access
	if payload.Role != "admin" {
		return fiber.NewError(
			fiber.StatusForbidden,
			"admin access required",
		)
	}

	// 3️⃣ Fetch users
	users, err := server.store.GetUsersByInstitute(
		c.Context(),
		payload.InstituteID,
	)
	if err != nil {
		return InternalServerError(err.Error())
	}

	// 4️⃣ Build safe response
	response := make([]fiber.Map, 0, len(users))

	for _, user := range users {
		role := ""
		if user.Role.Valid {
			role = user.Role.String
		}

		response = append(response, fiber.Map{
			"id":           user.ID,
			"institute_id": user.InstituteID,
			"name":         user.Name,
			"email":        user.Email,
			"role":         role,
			"is_active":    user.IsActive,
			"created_at":   user.CreatedAt,
		})
	}

	// 5️⃣ Return response
	return c.JSON(response)
}
