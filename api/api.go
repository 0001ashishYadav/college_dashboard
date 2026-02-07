package api

import (
	"dashboard/db/pgdb"
	"dashboard/token"
	"dashboard/utils"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type Server struct {
	app    *fiber.App
	store  pgdb.Store
	valid  *validator.Validate
	config utils.Config
	token  token.Maker
}

func NewServer(config utils.Config, store pgdb.Store, tokenMaker token.Maker) (*Server, error) {
	if store == nil {
		return nil, errors.New("store cannot be nil")
	}
	if tokenMaker == nil {
		return nil, errors.New("tokenMaker cannot be nil")
	}

	server := &Server{
		valid:  validator.New(),
		config: config,
		store:  store,
		token:  tokenMaker,
	}
	server.setupApi()
	return server, nil
}

func (server *Server) Start(port int16) error {
	return server.app.Listen(fmt.Sprintf(":%d", port))
}

type msgResponse struct {
	Msg string `json:"msg"`
}

func (server *Server) setupApi() {
	app := fiber.New(fiber.Config{
		ServerHeader:  "Inflection-Fiber",
		ErrorHandler:  errorHandler,
		BodyLimit:     2 * 1024 * 1024,
		CaseSensitive: true,
	})

	app.Use(logger.New(logger.ConfigDefault))

	app.Use(cors.New())

	app.Use(compress.New())

	// app.Use(csrf.New())

	app.Use(etag.New())

	app.Use(favicon.New())

	// app.Use(limiter.New())

	app.Get("/hello", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "hello"})
	})

	app.Post("/login", server.userLogin)
	app.Post("/user", server.authMiddleware, server.createUser)
	app.Put("/users/:id", server.authMiddleware, server.updateUser)
	app.Put("/users/:id/password", server.authMiddleware, server.UpdateUserPassword)
	app.Put("/users/:id/disable", server.authMiddleware, server.DisableUser)

	app.Get("/users/:id", server.authMiddleware, server.getUserByID)
	app.Get("/users", server.authMiddleware, server.getUserByEmail)
	app.Get("/institutes/users", server.authMiddleware, server.getUsersByInstitute)

	/////////////////////////////////   notice    ////////////////////////////////////////

	app.Post("/createNotice", server.authMiddleware, server.createNotice)
	app.Get("/notices/:id", server.authMiddleware, server.getNoticeByID)
	app.Get("/notices", server.authMiddleware, server.getNoticesByInstitute)

	app.Post("/notices/update/:id", server.authMiddleware, server.updateNotice)
	app.Post("/notices/:id/delete", server.authMiddleware, server.deleteNotice)

	/////////////////////////////////   photos    ////////////////////////////////////////

	app.Post("/photos", server.authMiddleware, server.createPhoto)
	app.Get("/photos/:id", server.authMiddleware, server.getPhotoByID)
	app.Get("/photos", server.authMiddleware, server.getPhotosByInstitute)
	app.Post("/photos/:id/image", server.authMiddleware, server.replacePhoto)

	server.app = app

}
