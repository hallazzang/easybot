package easybot

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// Server is an EasyBot server.
type Server struct {
	*fiber.App
	cfg ServerConfig
	db  *DB
}

// NewServer returns a new Server instance.
func NewServer(cfg ServerConfig, db *DB) *Server {
	server := &Server{
		App: fiber.New(cfg.Fiber),
		cfg: cfg,
		db:  db,
	}
	server.RouteV1()
	return server
}

// RouteV1 registers API v1 routes.
func (server *Server) RouteV1() {
	v1 := server.Group("/v1")

	bots := v1.Group("/bots")
	bots.Post("", server.CreateBot)
	bots.Get("/:id/messages", server.ReadBotMessages)
	bots.Post("/id:/messages", server.WriteBotMessages)
}

// CreateBot is a handler for creating a bot.
func (server *Server) CreateBot(c *fiber.Ctx) error {
	var body struct {
		Name        string `form:"name" json:"name"`
		Description string `form:"description" json:"description"`
	}
	if err := c.BodyParser(&body); err != nil {
		return err
	}
	if body.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	bot, err := server.db.CreateBot(context.TODO(), body.Name, body.Description)
	if err != nil {
		return fmt.Errorf("create bot: %w", err)
	}
	return c.JSON(bot)
}

// ReadBotMessages is a handler for reading bot messages.
func (server *Server) ReadBotMessages(c *fiber.Ctx) error {
	id := c.Params("id")
	_ = id
	var query struct {
		Peek bool `query:"peek"`
	}
	if err := c.QueryParser(&query); err != nil {
		return err
	}
	return c.JSON(fiber.Map{})
}

// WriteBotMessages is a handler for writing for messages.
func (server *Server) WriteBotMessages(c *fiber.Ctx) error {
	id := c.Params("id")
	_ = id
	return c.JSON(fiber.Map{})
}

// ErrorHandler is an error handler which returns JSON formatted error message.
func ErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	return c.Status(code).JSON(fiber.Map{"message": err.Error()})
}
