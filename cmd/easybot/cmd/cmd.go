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

func NewEasyBotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "easybot",
	}
	cmd.AddCommand(
		NewServeCmd(),
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
				accessKey = os.Getenv("EASYBOT_ACCESS_KEY")
			}
			if accessKey == "" {
				return fmt.Errorf("access key must be provided")
			}

			c, err := client.New(accessKey)
			if err != nil {
				return fmt.Errorf("new client: %w", err)
			}
			msgs, err := c.Room(botID, roomID).ReadMessages(peek)
			if err != nil {
				return fmt.Errorf("read messages: %w", err)
			}
			for _, msg := range msgs {
				fmt.Printf("[%s] %s\n", msg.CreatedAt.In(time.Local).Format(time.Kitchen), msg.Text)
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
				accessKey = os.Getenv("EASYBOT_ACCESS_KEY")
			}
			if accessKey == "" {
				return fmt.Errorf("access key must be provided")
			}

			c, err := client.New(accessKey)
			if err != nil {
				return fmt.Errorf("new client: %w", err)
			}
			if err := c.Room(botID, roomID).WriteMessages([]easybot.Message{{Text: text}}); err != nil {
				return fmt.Errorf("write messages: %w", err)
			}

			return nil
		},
	}
	cmd.Flags().StringVarP(&accessKey, "access-key", "k", "", "Access key")
	return cmd
}
