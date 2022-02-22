package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/hallazzang/easybot"
	"github.com/hallazzang/easybot/client"
)

const AccessKeyEnvKey = "EASYBOT_ACCESS_KEY"

func NewEasyBotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "easybot",
	}
	cmd.AddCommand(
		NewServeCmd(),
		NewCreateBotCmd(),
		NewListBotsCmd(),
		NewCreateRoomCmd(),
		NewListRoomsCmd(),
		NewReadCmd(),
		NewWriteCmd(),
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
			cfg := easybot.DefaultServerConfig // TODO: use ReadConfig

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

			c, err := client.New()
			if err != nil {
				return fmt.Errorf("new client: %w", err)
			}

			bot, err := c.CreateBot(name, desc)
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

			c, err := client.New()
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

			c, err := client.New()
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

			c, err := client.New()
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
	var (
		accessKey string
		peek      bool
	)
	cmd := &cobra.Command{
		Use:   "read [bot] [room]",
		Short: "Read messages",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			botID := args[0]
			roomID := args[1]
			if accessKey == "" {
				accessKey = os.Getenv(AccessKeyEnvKey)
			}
			if accessKey == "" {
				return fmt.Errorf("access key must be provided")
			}

			c, err := client.New(client.WithAccessKey(accessKey))
			if err != nil {
				return fmt.Errorf("new client: %w", err)
			}
			msgs, err := c.Room(botID, roomID).ReadMessages(context.TODO(), peek)
			if err != nil {
				return fmt.Errorf("read messages: %w", err)
			}
			fmt.Println("Created          Text")
			fmt.Println("---------------  ----")
			for _, msg := range msgs {
				fmt.Printf("%15s  %s\n", msg.CreatedAt.In(time.Local).Format(time.Stamp), msg.Text)
			}

			return nil
		},
	}
	cmd.Flags().StringVarP(&accessKey, "access-key", "k", "", "Access key")
	cmd.Flags().BoolVarP(&peek, "peek", "p", true, "Peek only")
	return cmd
}

func NewWriteCmd() *cobra.Command {
	var accessKey string
	cmd := &cobra.Command{
		Use:   "write [bot] [room] [text]",
		Args:  cobra.ExactArgs(3),
		Short: "Write messages",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			botID := args[0]
			roomID := args[1]
			text := args[2]
			if accessKey == "" {
				accessKey = os.Getenv(AccessKeyEnvKey)
			}
			if accessKey == "" {
				return fmt.Errorf("access key must be provided")
			}

			c, err := client.New(client.WithAccessKey(accessKey))
			if err != nil {
				return fmt.Errorf("new client: %w", err)
			}
			if err := c.Room(botID, roomID).WriteMessages(context.TODO(), []easybot.Message{{Text: text}}); err != nil {
				return fmt.Errorf("write messages: %w", err)
			}

			return nil
		},
	}
	cmd.Flags().StringVarP(&accessKey, "access-key", "k", "", "Access key")
	return cmd
}
