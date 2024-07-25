package main

import (
	"fmt"
	"log"
	"os"

	"http_test/apis"
	_ "http_test/core/config"

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
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func start() {
	fmt.Println("Starting the server....")
	apis.NewServer().StartAPI()
}
