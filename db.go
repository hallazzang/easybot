package easybot

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoBD collection names.
const (
	BotCollectionName     = "bots"
	RoomCollectionName    = "rooms"
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
		return Bot{}, fmt.Errorf("insert: %w", err)
	}
	bot.ID = ret.InsertedID.(primitive.ObjectID)
	return bot, nil
}

// GetBot returns a bot.
func (db *DB) GetBot(ctx context.Context, id primitive.ObjectID) (Bot, error) {
	coll := db.Database().Collection(BotCollectionName)
	var bot Bot
	if err := coll.FindOne(ctx, bson.M{IDKey: id}).Decode(&bot); err != nil {
		return Bot{}, fmt.Errorf("find: %w", err)
	}
	return bot, nil
}

// GetBots returns all bots.
// TODO: use pagination
func (db *DB) GetBots(ctx context.Context) ([]Bot, error) {
	coll := db.Database().Collection(BotCollectionName)
	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("find: %w", err)
	}
	var bots []Bot
	if err := cursor.All(ctx, &bots); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return bots, nil
}

// CreateRoom creates a new room.
func (db *DB) CreateRoom(ctx context.Context, botID primitive.ObjectID) (Room, error) {
	coll := db.Database().Collection(RoomCollectionName)
	room := Room{
		BotID:     botID,
		AccessKey: uuid.New().String(),
		CreatedAt: time.Now(),
	}
	ret, err := coll.InsertOne(ctx, room)
	if err != nil {
		return Room{}, fmt.Errorf("insert: %w", err)
	}
	room.ID = ret.InsertedID.(primitive.ObjectID)
	return room, nil
}

// GetRoom returns a room.
func (db *DB) GetRoom(ctx context.Context, id primitive.ObjectID) (Room, error) {
	coll := db.Database().Collection(RoomCollectionName)
	var room Room
	if err := coll.FindOne(ctx, bson.M{IDKey: id}).Decode(&room); err != nil {
		return Room{}, fmt.Errorf("find: %w", err)
	}
	return room, nil
}

// GetRooms returns all rooms.
// TODO: use pagination
func (db *DB) GetRooms(ctx context.Context, botID primitive.ObjectID) ([]Room, error) {
	coll := db.Database().Collection(RoomCollectionName)
	cursor, err := coll.Find(ctx, bson.M{RoomBotIDKey: botID})
	if err != nil {
		return nil, fmt.Errorf("find: %w", err)
	}
	var rooms []Room
	if err := cursor.All(ctx, &rooms); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return rooms, nil
}

// CreateMessages creates messages.
func (db *DB) CreateMessages(ctx context.Context, msgs []Message) ([]Message, error) {
	coll := db.Database().Collection(MessageCollectionName)
	var docs []interface{}
	for _, msg := range msgs {
		docs = append(docs, msg)
	}
	ret, err := coll.InsertMany(ctx, docs)
	if err != nil {
		return nil, fmt.Errorf("insert: %w", err)
	}
	res := make([]Message, len(msgs))
	for i, msg := range msgs {
		msg.ID = ret.InsertedIDs[i].(primitive.ObjectID)
		res[i] = msg
	}
	return res, nil
}

// GetUnreadMessages returns messages with specific type.
// TODO: use pagination
func (db *DB) GetUnreadMessages(ctx context.Context, roomID primitive.ObjectID, msgType MessageType) ([]Message, error) {
	coll := db.Database().Collection(MessageCollectionName)
	cursor, err := coll.Find(ctx, bson.M{
		MessageRoomIDKey: roomID,
		MessageTypeKey:   msgType,
		MessageReadKey:   false,
	})
	if err != nil {
		return nil, fmt.Errorf("find: %w", err)
	}
	var msgs []Message
	if err := cursor.All(ctx, &msgs); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return msgs, nil
}

// ReadMessages marks given messages as read.
// TODO: use pagination
func (db *DB) ReadMessages(ctx context.Context, msgs []Message) error {
	coll := db.Database().Collection(MessageCollectionName)
	var writes []mongo.WriteModel
	for _, msg := range msgs {
		writes = append(writes,
			mongo.NewUpdateOneModel().
				SetFilter(bson.M{IDKey: msg.ID}).
				SetUpdate(bson.M{"$set": bson.M{MessageReadKey: true}}))
	}
	if _, err := coll.BulkWrite(ctx, writes); err != nil {
		return fmt.Errorf("bulk write: %w", err)
	}
	return nil
}
