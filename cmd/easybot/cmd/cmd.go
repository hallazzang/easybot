package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/hallazzang/easybot"
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

			url := fmt.Sprintf("http://localhost:8000/v1/bots/%s/rooms/%s/messages", botID, roomID)
			if peek {
				url += "?peek=true"
			}
			req, _ := http.NewRequest("GET", url, nil)
			req.Header.Set(easybot.AccessKeyHeader, accessKey)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("http get: %w", err)
			}
			defer resp.Body.Close()

			var payload struct {
				Messages []struct {
					Text      string
					CreatedAt time.Time
				}
			}
			if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
				return fmt.Errorf("decode payload: %w", err)
			}
			for _, msg := range payload.Messages {
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

			url := fmt.Sprintf("http://localhost:8000/v1/bots/%s/rooms/%s/messages", botID, roomID)
			payload, _ := json.Marshal(map[string]interface{}{
				"messages": []easybot.Message{{Text: text}},
			})
			req, _ := http.NewRequest("POST", url, bytes.NewReader(payload))
			req.Header.Set(easybot.AccessKeyHeader, accessKey)
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("http get: %w", err)
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
		},
	}
	cmd.Flags().StringVarP(&accessKey, "access-key", "k", "", "Access key")
	return cmd
}
