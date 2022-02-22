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
	AccessKeyLocalsKey  = "accessKey"

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
	v1 := server.Group("/v1", server.AccessKeyMiddleware)

	bots := v1.Group("/bots")
	bots.Get("", server.ListBots)
	bots.Post("", server.CreateBot)

	bot := bots.Group("/:bot", server.BotMiddleware)
	bot.Get("/messages", server.ReadBotMessages)

	rooms := bot.Group("/rooms")
	rooms.Get("", server.ListRooms)
	rooms.Post("", server.CreateRoom)

	room := rooms.Group("/:room", server.RoomMiddleware, server.ClientTypeMiddleware)
	room.Get("/messages", server.ReadMessages)
	room.Post("/messages", server.WriteMessages)
}

type BotResponse struct {
	ID          primitive.ObjectID `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	AccessKey   string             `json:"accessKey,omitempty"`
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

// ListBots is a handler for listing all bots.
// TODO: use pagination
func (server *Server) ListBots(c *fiber.Ctx) error {
	bots, err := server.db.GetBots(context.TODO())
	if err != nil {
		return fmt.Errorf("get bots: %w", err)
	}
	resp := make([]BotResponse, len(bots))
	for i, bot := range bots {
		resp[i] = BotResponse{
			ID:          bot.ID,
			Name:        bot.Name,
			Description: bot.Description,
			CreatedAt:   bot.CreatedAt,
		}
	}
	return c.JSON(fiber.Map{
		"bots": resp,
	})
}

// ReadBotMessages is a handler for reading bot messages.
func (server *Server) ReadBotMessages(c *fiber.Ctx) error {
	bot := c.Locals(BotLocalsKey).(Bot)
	accessKey := c.Locals(AccessKeyLocalsKey).(string)
	var query struct {
		Peek bool `query:"peek"`
	}
	if err := c.QueryParser(&query); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if accessKey != bot.AccessKey {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	rooms, err := server.db.GetRooms(context.TODO(), bot.ID)
	if err != nil {
		return fmt.Errorf("get rooms: %w", err)
	}
	var msgs []Message
	var resp []MessageResponse
	for _, room := range rooms {
		ms, err := server.db.GetUnreadMessages(context.TODO(), room.ID, UserMessage)
		if err != nil {
			return fmt.Errorf("get unread messages: %w", err)
		}
		for _, msg := range ms {
			resp = append(resp, MessageResponse{
				ID:        msg.ID,
				RoomID:    msg.RoomID,
				Type:      msg.Type,
				Text:      msg.Text,
				CreatedAt: msg.CreatedAt,
			})
		}
		msgs = append(msgs, ms...)
	}
	if !query.Peek && len(msgs) > 0 {
		if err := server.db.ReadMessages(context.TODO(), msgs); err != nil {
			return fmt.Errorf("read messages: %w", err)
		}
	}
	return c.JSON(fiber.Map{
		"messages": resp,
	})
}

type RoomResponse struct {
	ID        primitive.ObjectID `json:"id"`
	BotID     primitive.ObjectID `json:"botID"`
	AccessKey string             `json:"accessKey,omitempty"`
	CreatedAt time.Time          `json:"createdAt"`
}

// CreateRoom is a handler for creating a room.
func (server *Server) CreateRoom(c *fiber.Ctx) error {
	bot := c.Locals(BotLocalsKey).(Bot)
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

// ListRooms is a handler for listing all rooms.
func (server *Server) ListRooms(c *fiber.Ctx) error {
	bot := c.Locals(BotLocalsKey).(Bot)
	rooms, err := server.db.GetRooms(context.TODO(), bot.ID)
	if err != nil {
		return fmt.Errorf("get rooms: %w", err)
	}
	resp := make([]RoomResponse, len(rooms))
	for i, room := range rooms {
		resp[i] = RoomResponse{
			ID:        room.ID,
			BotID:     room.BotID,
			CreatedAt: room.CreatedAt,
		}
	}
	return c.JSON(fiber.Map{
		"rooms": resp,
	})
}

type ClientType string

const (
	BotClient  = ClientType("bot")
	UserClient = ClientType("user")
)

type MessageRequest struct {
	RoomID primitive.ObjectID `json:"roomID"`
	Text   string             `json:"text"`
}

type MessageResponse struct {
	ID        primitive.ObjectID `json:"id"`
	RoomID    primitive.ObjectID `json:"roomID"`
	Type      MessageType        `json:"type"`
	Text      string             `json:"text"`
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
	if clientType == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	var msgType MessageType
	switch clientType {
	case BotClient:
		msgType = UserMessage
	case UserClient:
		msgType = BotMessage
	}
	msgs, err := server.db.GetUnreadMessages(context.TODO(), room.ID, msgType)
	if err != nil {
		return fmt.Errorf("get unread messages: %w", err)
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
		Messages []MessageRequest `json:"messages"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	room := c.Locals(RoomLocalsKey).(Room)
	clientType := c.Locals(ClientTypeLocalsKey).(ClientType)
	if clientType == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	now := time.Now()
	msgs := make([]Message, len(body.Messages))
	for i := range body.Messages {
		var msgType MessageType
		switch clientType {
		case BotClient:
			msgType = BotMessage
		case UserClient:
			msgType = UserMessage
		}
		msgs[i] = Message{
			RoomID:    room.ID,
			Type:      msgType,
			Text:      body.Messages[i].Text,
			CreatedAt: now,
		}
	}
	msgs, err := server.db.CreateMessages(context.TODO(), msgs)
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
			CreatedAt: msg.CreatedAt,
		}
	}
	return c.JSON(fiber.Map{
		"messages": resp,
	})
}

func (server *Server) AccessKeyMiddleware(c *fiber.Ctx) error {
	var hdr struct {
		AccessKey string `reqHeader:"X-Access-Key"`
	}
	if err := c.ReqHeaderParser(&hdr); err != nil {
		return fmt.Errorf("read req header: %w", err)
	}
	c.Locals(AccessKeyLocalsKey, hdr.AccessKey)
	return c.Next()
}

func (server *Server) BotMiddleware(c *fiber.Ctx) error {
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
	c.Locals(BotLocalsKey, bot)
	return c.Next()
}

// RoomMiddleware is a middleware for a room.
func (server *Server) RoomMiddleware(c *fiber.Ctx) error {
	bot := c.Locals(BotLocalsKey).(Bot)
	roomID, err := primitive.ObjectIDFromHex(c.Params("room"))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("room %s not found", c.Params("room")))
	}
	room, err := server.db.GetRoom(context.TODO(), roomID)
	if err != nil || room.BotID != bot.ID {
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("get room: %w", err)
		}
		return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("room %s not found", roomID))
	}
	c.Locals(RoomLocalsKey, room)
	return c.Next()
}

func (server *Server) ClientTypeMiddleware(c *fiber.Ctx) error {
	accessKey := c.Locals(AccessKeyLocalsKey).(string)
	bot := c.Locals(BotLocalsKey).(Bot)
	room := c.Locals(RoomLocalsKey).(Room)
	var clientType ClientType
	if accessKey == bot.AccessKey {
		clientType = BotClient
	} else if accessKey == room.AccessKey {
		clientType = UserClient
	}
	c.Locals(ClientTypeLocalsKey, clientType)
	return c.Next()
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
