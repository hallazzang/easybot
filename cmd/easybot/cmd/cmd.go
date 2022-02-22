package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hallazzang/read"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/hallazzang/easybot"
	"github.com/hallazzang/easybot/client"
)

func NewEasyBotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "easybot",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			viper.SetConfigName("easybot")
			viper.AddConfigPath(".")
			home, err := os.UserHomeDir()
			if err == nil {
				viper.AddConfigPath(filepath.Join(home, ".easybot"))
			}
			if err := viper.ReadInConfig(); err != nil {
				if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
					return fmt.Errorf("read config: %w", err)
				}
			}
			return nil
		},
	}
	cmd.AddCommand(
		NewServeCmd(),
		NewCreateBotCmd(),
		NewListBotsCmd(),
		NewCreateRoomCmd(),
		NewListRoomsCmd(),
		NewReadCmd(),
		NewWriteCmd(),
		NewInteractCmd(),
	)
	return cmd
}

func NewServeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "serve [addr]",
		Short:   "Run an EasyBot server",
		Aliases: []string{"s"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			addr := args[0]
			cfg := easybot.DefaultServerConfig
			if err := viper.UnmarshalKey("server", &cfg); err != nil {
				return fmt.Errorf("unmarshal server config: %w", err)
			}

			db, err := easybot.NewDB(context.Background(), cfg.DB)
			if err != nil {
				return fmt.Errorf("new db: %w", err)
			}
			defer db.Close()

			server := easybot.NewServer(cfg, db)

			if err := server.Listen(addr); err != nil {
				return fmt.Errorf("listen: %w", err)
			}

			return nil
		},
	}
	return cmd
}

func NewCreateBotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-bot [name] [description]",
		Short: "Create a bot",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			name := args[0]
			var desc string
			if len(args) > 1 {
				desc = args[1]
			}

			cfg := client.DefaultConfig
			if err := viper.UnmarshalKey("client", &cfg); err != nil {
				return fmt.Errorf("unmarshal client config: %w", err)
			}

			c, err := client.New(cfg)
			if err != nil {
				return fmt.Errorf("new client: %w", err)
			}

			bot, err := c.CreateBot(context.TODO(), name, desc)
			if err != nil {
				return fmt.Errorf("create bot: %w", err)
			}

			fmt.Printf("bot id: %s\naccess key: %s\n", bot.ID, bot.AccessKey)
			return nil
		},
	}
	return cmd
}

func NewListBotsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bots",
		Short: "List all bots",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			cfg := client.DefaultConfig
			if err := viper.UnmarshalKey("client", &cfg); err != nil {
				return fmt.Errorf("unmarshal client config: %w", err)
			}

			c, err := client.New(cfg)
			if err != nil {
				return fmt.Errorf("new client: %w", err)
			}

			bots, err := c.ListBots(context.TODO())
			if err != nil {
				return fmt.Errorf("list bots: %w", err)
			}

			fmt.Println("ID                        Created          Name")
			fmt.Println("------------------------  ---------------  ----")
			for _, bot := range bots {
				fmt.Printf("%24s  %15s  %s\n", bot.ID.Hex(), bot.CreatedAt.In(time.Local).Format(time.Stamp), bot.Name)
			}
			return nil
		},
	}
	return cmd
}

func NewCreateRoomCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-room [bot]",
		Short: "Create a room",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			botID := args[0]

			cfg := client.DefaultConfig
			if err := viper.UnmarshalKey("client", &cfg); err != nil {
				return fmt.Errorf("unmarshal client config: %w", err)
			}

			c, err := client.New(cfg)
			if err != nil {
				return fmt.Errorf("new client: %w", err)
			}

			room, err := c.CreateRoom(context.TODO(), botID)
			if err != nil {
				return fmt.Errorf("create room: %w", err)
			}

			fmt.Printf("room id: %s\naccess key: %s\n", room.ID, room.AccessKey)
			return nil
		},
	}
	return cmd
}

func NewListRoomsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rooms [bot]",
		Short: "List all rooms of a bot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			botID := args[0]

			cfg := client.DefaultConfig
			if err := viper.UnmarshalKey("client", &cfg); err != nil {
				return fmt.Errorf("unmarshal client config: %w", err)
			}

			c, err := client.New(cfg)
			if err != nil {
				return fmt.Errorf("new client: %w", err)
			}

			rooms, err := c.ListRooms(context.TODO(), botID)
			if err != nil {
				return fmt.Errorf("list rooms: %w", err)
			}

			fmt.Println("ID                        Created   ")
			fmt.Println("------------------------  ----------")
			for _, room := range rooms {
				fmt.Printf("%24s  %s\n", room.ID.Hex(), room.CreatedAt.In(time.Local).Format(time.Stamp))
			}

			return nil
		},
	}
	return cmd
}

func NewReadCmd() *cobra.Command {
	var peek bool
	cmd := &cobra.Command{
		Use:   "read [bot] [room]",
		Short: "Read messages",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			botID := args[0]
			var roomID string
			if len(args) > 1 {
				roomID = args[1]
			}

			cfg := client.DefaultConfig
			if err := viper.UnmarshalKey("client", &cfg); err != nil {
				return fmt.Errorf("unmarshal client config: %w", err)
			}

			c, err := client.New(cfg)
			if err != nil {
				return fmt.Errorf("new client: %w", err)
			}
			var msgs []easybot.MessageResponse
			if roomID == "" {
				msgs, err = c.Bot(botID).ReadMessages(context.TODO(), peek)
				if err != nil {
					return fmt.Errorf("read messages: %w", err)
				}
				fmt.Println("Room                      Created          Text")
				fmt.Println("------------------------  ---------------  ----")
				for _, msg := range msgs {
					fmt.Printf("%24s  %15s  %s\n", msg.RoomID.Hex(), msg.CreatedAt.In(time.Local).Format(time.Stamp), msg.Text)
				}
			} else {
				msgs, err = c.Room(botID, roomID).ReadMessages(context.TODO(), peek)
				if err != nil {
					return fmt.Errorf("read messages: %w", err)
				}
				fmt.Println("Created          Text")
				fmt.Println("---------------  ----")
				for _, msg := range msgs {
					fmt.Printf("%15s  %s\n", msg.CreatedAt.In(time.Local).Format(time.Stamp), msg.Text)
				}
			}

			return nil
		},
	}
	cmd.Flags().BoolVarP(&peek, "peek", "p", true, "Peek only")
	return cmd
}

func NewWriteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "write [bot] [room] [text]",
		Args:  cobra.ExactArgs(3),
		Short: "Write messages",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			botID := args[0]
			roomID := args[1]
			text := args[2]

			cfg := client.DefaultConfig
			if err := viper.UnmarshalKey("client", &cfg); err != nil {
				return fmt.Errorf("unmarshal client config: %w", err)
			}

			c, err := client.New(cfg)
			if err != nil {
				return fmt.Errorf("new client: %w", err)
			}
			if err := c.Room(botID, roomID).WriteMessages(context.TODO(), []easybot.MessageRequest{{Text: text}}); err != nil {
				return fmt.Errorf("write messages: %w", err)
			}

			return nil
		},
	}
	return cmd
}

func NewInteractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "interact [bot] [room]",
		Short: "Interact within a room",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			botID := args[0]
			roomID := args[1]

			cfg := client.DefaultConfig
			if err := viper.UnmarshalKey("client", &cfg); err != nil {
				return fmt.Errorf("unmarshal client config: %w", err)
			}

			c, err := client.New(cfg)
			if err != nil {
				return fmt.Errorf("access key must be provided")
			}

			room := c.Room(botID, roomID)
			for {
				fmt.Print("text> ")
				text, err := read.Line()
				if err != nil {
					return fmt.Errorf("read line: %w", err)
				}

				if err := room.WriteMessages(context.TODO(), []easybot.MessageRequest{
					{Text: text},
				}); err != nil {
					return fmt.Errorf("write messages: %w", err)
				}

				for {
					time.Sleep(100 * time.Millisecond)

					msgs, err := room.ReadMessages(context.TODO(), false)
					if err != nil {
						return fmt.Errorf("read messages: %w", err)
					}

					received := false
					for _, msg := range msgs {
						fmt.Printf("received: %s\n", msg.Text)
						received = true
					}
					if received {
						break
					}
				}
			}
		},
	}
	return cmd
}
