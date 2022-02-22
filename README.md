# easybot

Easy chatbot coding for education purpose.

## Example

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/hallazzang/easybot"
	"github.com/hallazzang/easybot/client"
)

const (
	ServerURL = "<server-url>"
	BotID     = "<bot-id>"
	AccessKey = "<bot-access-key>"
)

func main() {
	c, err := client.New(client.Config{
		ServerURL: ServerURL,
		AccessKey: AccessKey,
	})
	if err != nil {
		panic(err)
	}

	bot := c.Bot(BotID)

	fmt.Printf("bot id: %s\nbot access key: %s\n", bot.ID, bot.AccessKey)

	for {
		msgs, err := bot.ReadMessages(context.TODO(), false)
		if err != nil {
			panic(err)
		}

		for _, msg := range msgs {
			fmt.Printf("text: %s\n", msg.Text)
			bot.Room(msg.RoomID.Hex()).WriteMessages(context.TODO(), []easybot.MessageRequest{
				{Text: "You said, " + msg.Text},
			})
		}

		time.Sleep(100 * time.Millisecond)
	}
}
```
