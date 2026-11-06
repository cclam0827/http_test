package main

import (
	"log/slog"
	"os"

	"http_test/apis"
	"http_test/core/config"
	"http_test/core/logger"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start the http server",
				Action: func(c *cli.Context) error {
					start()
					return nil
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
}

func start() {
	cfg := config.GetConfig()
	logger.Setup(cfg.App.LogFormat, cfg.App.LogLevel, cfg.App.Environment)
	apis.NewServer().StartAPI()
}
