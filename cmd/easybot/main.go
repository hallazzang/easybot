package main

import (
	"os"

	"github.com/hallazzang/easybot/cmd/easybot/cmd"
)

func main() {
	if err := cmd.NewEasyBotCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
