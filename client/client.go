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
	ServerURL string
}

type Option func(cfg *Config) error

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

func New(accessKey string, opts ...Option) (*Client, error) {
	c := &Client{
		accessKey:  accessKey,
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
	if cfg.ServerURL != "" {
		u, err := url.Parse(cfg.ServerURL)
		if err != nil {
			return nil, fmt.Errorf("parse server url: %w", err)
		}
		c.serverURL = u
	}
	return c, nil
}

func (c *Client) Room(botID, roomID string) *Room {
	return &Room{c: c, botID: botID, roomID: roomID}
}

type Room struct {
	c      *Client
	botID  string
	roomID string
}

func (room *Room) ReadMessages(peek bool) ([]easybot.Message, error) {
	u, _ := room.c.serverURL.Parse(fmt.Sprintf("bots/%s/rooms/%s/messages", room.botID, room.roomID))
	if peek {
		u.RawQuery = url.Values{"peek": {"true"}}.Encode()
	}
	req, _ := http.NewRequest("GET", u.String(), nil)
	req.Header.Set(easybot.HeaderAccessKey, room.c.accessKey)
	resp, err := room.c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()
	var payload struct {
		Messages []easybot.Message
	}
	if resp.StatusCode != http.StatusOK {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read body: %w", err)
		}
		return nil, fmt.Errorf("error: %s", data)
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode body: %w", err)
	}
	return payload.Messages, nil
}

func (room *Room) WriteMessages(msgs []easybot.Message) error {
	payload, _ := json.Marshal(map[string]interface{}{"messages": msgs})
	u, _ := room.c.serverURL.Parse(fmt.Sprintf("bots/%s/rooms/%s/messages", room.botID, room.roomID))
	req, _ := http.NewRequest("POST", u.String(), bytes.NewReader(payload))
	req.Header.Set(easybot.HeaderAccessKey, room.c.accessKey)
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
		return fmt.Errorf("error: %s", data)
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}
