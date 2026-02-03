package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type adminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginAdminResponse struct {
	Token           string    `json:"token"`
	TokenExperiesAt time.Time `json:"token_experies_at"`
	Name            string    `json:"name"`
	Role            string    `json:"role"`
	Email           string    `json:"email"`
}

func (server *Server) login(c *fiber.Ctx) error {
	var req adminLoginRequest
	if err := c.BodyParser(&req); err != nil {
		return err
	}
	validationError := server.validate(req)
	if validationError != nil {

	}
}
