package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/hallazzang/easybot"
)

func NewEasyBotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "easybot",
	}
	cmd.AddCommand(
		NewServeCmd(),
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
			cfg := easybot.DefaultServerConfig // TODO: use ReadConfig

			db, err := easybot.NewDB(context.Background(), cfg.DB)
			if err != nil {
				return fmt.Errorf("new db: %w", err)
			}
			defer db.Close()

			server := easybot.NewServer(cfg, db)

			if err := server.Listen(args[0]); err != nil {
				return fmt.Errorf("listen: %w", err)
			}

			return nil
		},
	}
	return cmd
}
