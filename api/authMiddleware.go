package api

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const (
	AuthorizationHeaderKey = "Authorization"
	BearerPrefix           = "Bearer "
	TokenKey               = "token"
	TokenPayloadKey        = "token_payload"
)

func (server *Server) authMiddleware(c *fiber.Ctx) error {

	fmt.Println("************** AuthMiddleware ******************")

	var accessToken string

	// 1️⃣ Read Authorization header
	authHeader := c.Get(AuthorizationHeaderKey)

	if strings.HasPrefix(authHeader, BearerPrefix) {
		accessToken = strings.TrimPrefix(authHeader, BearerPrefix)
	}

	// 2️⃣ Fallback: read token from query param
	if accessToken == "" {
		accessToken = c.Query(TokenKey)
	}

	// 3️⃣ Token missing
	if accessToken == "" {
		return &fiber.Error{
			Code:    fiber.StatusUnauthorized,
			Message: "authorization token is required",
		}
	}

	// 4️⃣ Verify token
	payload, err := server.token.VerifyToken(accessToken)
	if err != nil {
		return &fiber.Error{
			Code:    fiber.StatusUnauthorized,
			Message: "invalid or expired token",
		}
	}

	// 5️⃣ Store payload data in context
	c.Locals("user_id", payload.ID)
	c.Locals("email", payload.Email)
	c.Locals("role", payload.Role)
	c.Locals("name", payload.Name)
	c.Locals(TokenPayloadKey, payload)

	// 6️⃣ Continue request
	return c.Next()
}
