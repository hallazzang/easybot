package client

const DefaultServerURL = "http://localhost:8000"

var DefaultConfig = Config{
	ServerURL: DefaultServerURL,
}

type Config struct {
	ServerURL string
	AccessKey string
}
