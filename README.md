# easybot

Easy chatbot coding for education purpose.

## Installation

```
go install github.com/hallazzang/easybot/cmd/easybot@v0.2.2
```

## Example

### Bot

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

### Client

Create a file named `easybot.yml` in `~/.easybot` directory(or, you can just create the file inside the current directory, too):
```yaml
Client:
  ServerURL: <server-url>
```

First, create a room:
```
$ easybot create-room <bot-id>
room id: <room-id>
access key: <user-access-key>
```

Remembee the `<room-id>`. This will be used to interact with the bot. Copy `<user-access-key>` part and paste it in the config file `easybot.yml`:
```yaml
Client:
  ServerURL: <server-url>
  AccessKey: <user-access-key>
```

Now you can interact with the bot:
```
$ easybot interact <bot-id> <room-id>
text> Hello
received: You said, Hello
text>
```
