package easybot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type LocalsKey string

const (
	BotLocalsKey        = "bot"
	RoomLocalsKey       = "room"
	ClientTypeLocalsKey = "clientType"

	HeaderAccessKey = "X-Access-Key"
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

	bot := bots.Group("/:bot")
	//bot.Get("/messages", server.ReadBotMessages)

	rooms := bot.Group("/rooms")
	rooms.Post("", server.CreateRoom)
	rooms.Get("/:room/messages", server.RoomMiddleware, server.ReadMessages)
	rooms.Post("/:room/messages", server.RoomMiddleware, server.WriteMessages)
}

type BotResponse struct {
	ID          primitive.ObjectID `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	AccessKey   string             `json:"accessKey"`
	CreatedAt   time.Time          `json:"createdAt"`
}

// CreateBot is a handler for creating a bot.
func (server *Server) CreateBot(c *fiber.Ctx) error {
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if body.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	bot, err := server.db.CreateBot(context.TODO(), body.Name, body.Description)
	if err != nil {
		return fmt.Errorf("create bot: %w", err)
	}
	return c.JSON(BotResponse{
		ID:          bot.ID,
		Name:        bot.Name,
		Description: bot.Description,
		AccessKey:   bot.AccessKey,
		CreatedAt:   bot.CreatedAt,
	})
}

type RoomResponse struct {
	ID        primitive.ObjectID `json:"id"`
	BotID     primitive.ObjectID `json:"botID"`
	AccessKey string             `json:"accessKey"`
	CreatedAt time.Time          `json:"createdAt"`
}

// CreateRoom is a handler for creating a room.
func (server *Server) CreateRoom(c *fiber.Ctx) error {
	botID, err := primitive.ObjectIDFromHex(c.Params("bot"))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("bot %s not found", c.Params("bot")))
	}
	bot, err := server.db.GetBot(context.TODO(), botID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("bot %s not found", botID))
		}
		return fmt.Errorf("get bot: %w", err)
	}
	room, err := server.db.CreateRoom(context.TODO(), bot.ID)
	if err != nil {
		return fmt.Errorf("create room: %w", err)
	}
	return c.JSON(RoomResponse{
		ID:        room.ID,
		BotID:     room.BotID,
		AccessKey: room.AccessKey,
		CreatedAt: room.CreatedAt,
	})
}

type ClientType string

const (
	BotClient  = ClientType("bot")
	UserClient = ClientType("user")
)

// RoomMiddleware is a middleware for a room.
func (server *Server) RoomMiddleware(c *fiber.Ctx) error {
	botID, err := primitive.ObjectIDFromHex(c.Params("bot"))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("bot %s not found", c.Params("bot")))
	}
	roomID, err := primitive.ObjectIDFromHex(c.Params("room"))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("room %s not found", c.Params("room")))
	}
	var hdr struct {
		AccessKey string `reqHeader:"X-Access-Key"`
	}
	if err := c.ReqHeaderParser(&hdr); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	bot, err := server.db.GetBot(context.TODO(), botID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("bot %s not found", botID))
		}
		return fmt.Errorf("get bot: %w", err)
	}
	var clientType ClientType
	if hdr.AccessKey == bot.AccessKey {
		clientType = BotClient
	}
	room, err := server.db.GetRoom(context.TODO(), roomID)
	if err != nil || room.BotID != bot.ID {
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("get room: %w", err)
		}
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("room %s not found", roomID))
	}
	if hdr.AccessKey == room.AccessKey {
		clientType = UserClient
	}
	if clientType == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	c.Locals(BotLocalsKey, bot)
	c.Locals(RoomLocalsKey, room)
	c.Locals(ClientTypeLocalsKey, clientType)
	return c.Next()
}

type MessageResponse struct {
	ID        primitive.ObjectID `json:"id"`
	RoomID    primitive.ObjectID `json:"roomID"`
	Type      MessageType        `json:"type"`
	ReplyTo   primitive.ObjectID `json:"replyTo,omitempty"`
	Text      string             `json:"text"`
	Read      bool               `json:"read"`
	CreatedAt time.Time          `json:"createdAt"`
}

// ReadMessages is a handler for reading messages in a room.
func (server *Server) ReadMessages(c *fiber.Ctx) error {
	var query struct {
		Peek bool `query:"peek"`
	}
	if err := c.QueryParser(&query); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	room := c.Locals(RoomLocalsKey).(Room)
	clientType := c.Locals(ClientTypeLocalsKey).(ClientType)
	var msgType MessageType
	switch clientType {
	case BotClient:
		msgType = UserMessage
	case UserClient:
		msgType = BotMessage
	}
	msgs, err := server.db.GetUnreadMessages(context.TODO(), room.ID, msgType)
	if err != nil {
		return fmt.Errorf("get messages: %w", err)
	}
	if !query.Peek && len(msgs) > 0 {
		if err := server.db.ReadMessages(context.TODO(), msgs); err != nil {
			return fmt.Errorf("read messages: %w", err)
		}
	}
	resp := make([]MessageResponse, len(msgs))
	for i, msg := range msgs {
		resp[i] = MessageResponse{
			ID:        msg.ID,
			RoomID:    msg.RoomID,
			Type:      msg.Type,
			Text:      msg.Text,
			Read:      msg.Read,
			CreatedAt: msg.CreatedAt,
		}
	}
	return c.JSON(fiber.Map{
		"messages": resp,
	})
}

// WriteMessages is a handler for writing messages in a room.
func (server *Server) WriteMessages(c *fiber.Ctx) error {
	var body struct {
		Messages []Message `json:"messages"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	room := c.Locals(RoomLocalsKey).(Room)
	clientType := c.Locals(ClientTypeLocalsKey).(ClientType)
	now := time.Now()
	for i := range body.Messages {
		var msgType MessageType
		switch clientType {
		case BotClient:
			msgType = BotMessage
		case UserClient:
			msgType = UserMessage
		}
		body.Messages[i] = Message{
			RoomID:    room.ID,
			Type:      msgType,
			Text:      body.Messages[i].Text,
			CreatedAt: now,
		}
	}
	msgs, err := server.db.CreateMessages(context.TODO(), body.Messages)
	if err != nil {
		return fmt.Errorf("create messages: %w", err)
	}
	resp := make([]MessageResponse, len(msgs))
	for i, msg := range msgs {
		resp[i] = MessageResponse{
			ID:        msg.ID,
			RoomID:    msg.RoomID,
			Type:      msg.Type,
			Text:      msg.Text,
			Read:      msg.Read,
			CreatedAt: msg.CreatedAt,
		}
	}
	return c.JSON(fiber.Map{
		"messages": resp,
	})
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
