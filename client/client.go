package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/hallazzang/easybot"
)

const DefaultServerURL = "http://localhost:8000/v1/"

type Config struct {
	AccessKey string
	ServerURL string
}

type Option func(cfg *Config) error

func WithAccessKey(accessKey string) Option {
	return func(cfg *Config) error {
		cfg.AccessKey = accessKey
		return nil
	}
}

func WithServerURL(serverURL string) Option {
	return func(cfg *Config) error {
		cfg.ServerURL = serverURL
		return nil
	}
}

type Client struct {
	accessKey  string
	serverURL  *url.URL
	httpClient *http.Client
}

func New(opts ...Option) (*Client, error) {
	c := &Client{
		httpClient: &http.Client{},
	}
	cfg := Config{
		ServerURL: DefaultServerURL,
	}
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return nil, err
		}
	}
	c.accessKey = cfg.AccessKey
	if cfg.ServerURL != "" {
		u, err := url.Parse(cfg.ServerURL)
		if err != nil {
			return nil, fmt.Errorf("parse server url: %w", err)
		}
		c.serverURL = u
	}
	return c, nil
}

type Bot struct {
	c         *Client
	AccessKey string
	ID        string
}

func (c *Client) CreateBot(name, description string) (*Bot, error) {
	u, _ := c.serverURL.Parse("bots")
	payload, _ := json.Marshal(map[string]interface{}{
		"name":        name,
		"description": description,
	})
	req, _ := http.NewRequest("POST", u.String(), bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read body: %w", err)
		}
		return nil, fmt.Errorf("bad status code: %d: %s", resp.StatusCode, data)
	}
	var body easybot.BotResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode body: %w", err)
	}
	return &Bot{c: c, AccessKey: body.AccessKey, ID: body.ID.Hex()}, nil
}

func (c *Client) Bot(id string) *Bot {
	return &Bot{
		c:         c,
		AccessKey: c.accessKey,
		ID:        id,
	}
}

func (bot *Bot) Room(roomID string) *Room {
	return &Room{c: bot.c, AccessKey: bot.AccessKey, BotID: bot.ID, ID: roomID}
}

type Room struct {
	c         *Client
	AccessKey string
	BotID     string
	ID        string
}

func (c *Client) CreateRoom(botID string) (*Room, error) {
	u, _ := c.serverURL.Parse(fmt.Sprintf("bots/%s/rooms", botID))
	resp, err := c.httpClient.Post(u.String(), "", nil)
	if err != nil {
		return nil, fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read body: %w", err)
		}
		return nil, fmt.Errorf("bad status code: %d: %s", resp.StatusCode, data)
	}
	var body easybot.RoomResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode body: %w", err)
	}
	return &Room{c: c, AccessKey: body.AccessKey, BotID: botID, ID: body.ID.Hex()}, nil
}

func (c *Client) Room(botID, id string) *Room {
	return &Room{c: c, AccessKey: c.accessKey, BotID: botID, ID: id}
}

func (room *Room) ReadMessages(peek bool) ([]easybot.Message, error) {
	u, _ := room.c.serverURL.Parse(fmt.Sprintf("bots/%s/rooms/%s/messages", room.BotID, room.ID))
	if peek {
		u.RawQuery = url.Values{"peek": {"true"}}.Encode()
	}
	req, _ := http.NewRequest("GET", u.String(), nil)
	req.Header.Set(easybot.HeaderAccessKey, room.AccessKey)
	resp, err := room.c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()
	var body struct {
		Messages []easybot.Message
	}
	if resp.StatusCode != http.StatusOK {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read body: %w", err)
		}
		return nil, fmt.Errorf("bad status code: %d: %s", resp.StatusCode, data)
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode body: %w", err)
	}
	return body.Messages, nil
}

func (room *Room) WriteMessages(msgs []easybot.Message) error {
	payload, _ := json.Marshal(map[string]interface{}{"messages": msgs})
	u, _ := room.c.serverURL.Parse(fmt.Sprintf("bots/%s/rooms/%s/messages", room.BotID, room.ID))
	req, _ := http.NewRequest("POST", u.String(), bytes.NewReader(payload))
	req.Header.Set(easybot.HeaderAccessKey, room.AccessKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := room.c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read body: %w", err)
		}
		return fmt.Errorf("bad status code: %d: %s", resp.StatusCode, data)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}
