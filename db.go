package easybot

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoBD collection names.
const (
	BotCollectionName     = "bots"
	MessageCollectionName = "messages"
)

// DB is the database interface.
type DB struct {
	cfg         DBConfig
	mongoClient *mongo.Client
}

// NewDB connects to the mongodb server and returns a new DB instance.
func NewDB(ctx context.Context, cfg DBConfig) (*DB, error) {
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	return &DB{
		cfg:         cfg,
		mongoClient: mongoClient,
	}, nil
}

// Close disconnects from the mongodb server.
func (db *DB) Close() error {
	return db.mongoClient.Disconnect(context.TODO())
}

// Database returns the mongodb database.
func (db *DB) Database() *mongo.Database {
	return db.mongoClient.Database(db.cfg.Database)
}

// CreateBot creates a new bot.
func (db *DB) CreateBot(ctx context.Context, name, desc string) (Bot, error) {
	coll := db.Database().Collection(BotCollectionName)
	bot := Bot{
		Name:        name,
		Description: desc,
		AccessKey:   uuid.New().String(),
		CreatedAt:   time.Now(),
	}
	ret, err := coll.InsertOne(ctx, bot)
	if err != nil {
		return Bot{}, fmt.Errorf("insert one: %w", err)
	}
	bot.ID = ret.InsertedID.(primitive.ObjectID)
	return bot, nil
}
