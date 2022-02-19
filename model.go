package easybot

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Bot is the model for a bot.
type Bot struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	AccessKey   string             `bson:"accessKey" json:"accessKey"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
}

// Message is the model for a message.
type Message struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`
}
