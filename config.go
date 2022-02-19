package easybot

import (
	"github.com/gofiber/fiber/v2"
)

var (
	DefaultServerConfig = ServerConfig{
		Fiber: fiber.Config{
			ErrorHandler: ErrorHandler,
		},
		DB: DefaultDBConfig,
	}

	DefaultDBConfig = DBConfig{
		URI:      "mongodb://localhost",
		Database: "easybot",
	}
)

type ServerConfig struct {
	Fiber fiber.Config
	DB    DBConfig
}

type DBConfig struct {
	URI      string
	Database string
}

// TODO: create ReadConfig()
