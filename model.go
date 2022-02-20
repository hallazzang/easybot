package easybot

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Common key names.
const (
	IDKey        = "_id"
	CreatedAtKey = "createdAt"
)

// Bot key names.
const (
	BotNameKey        = "name"
	BotDescriptionKey = "description"
	BotAccessKeyKey   = "accessKey"
)

// Bot is the model for a bot.
type Bot struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Description string             `bson:"description"`
	AccessKey   string             `bson:"accessKey"` // access key of a bot.
	CreatedAt   time.Time          `bson:"createdAt"`
}

// Room key names.
const (
	RoomBotIDKey     = "botID"
	RoomAccessKeyKey = "accessKey"
)

// Room is the model for a room.
type Room struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	BotID     primitive.ObjectID `bson:"botID"`
	AccessKey string             `bson:"accessKey"` // access key of a user.
	CreatedAt time.Time          `bson:"createdAt"`
}

type MessageType string

// MessageType enumerations.
const (
	BotMessage  = MessageType("bot_message")
	UserMessage = MessageType("user_message")
)

// Message key names.
const (
	MessageRoomIDKey  = "roomID"
	MessageTypeKey    = "type"
	MessageReplyToKey = "replyTo"
	MessageTextKey    = "text"
	MessageReadKey    = "read"
)

// Message is the model for a message.
type Message struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	RoomID    primitive.ObjectID `bson:"roomID"`
	Type      MessageType        `bson:"type"`
	ReplyTo   primitive.ObjectID `bson:"replyTo,omitempty"`
	Text      string             `bson:"text"`
	Read      bool               `bson:"read"`
	CreatedAt time.Time          `bson:"createdAt"`
}
