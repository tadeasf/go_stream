package main

import (
	"log"
	"os"

	"github.com/tadeasf/go_stream/src/stream"
	"github.com/urfave/cli/v2"
)

const version = "1.0.0"

func main() {
	app := &cli.App{
		Name:    "go_stream",
		Usage:   "A high-performance video streaming server",
		Version: version,
		Commands: []*cli.Command{
			{
				Name:  "serve",
				Usage: "Start the streaming server",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "recursive",
						Aliases: []string{"r"},
						Usage:   "Search for videos recursively",
					},
					&cli.IntFlag{
						Name:    "port",
						Aliases: []string{"p"},
						Value:   8069,
						Usage:   "Port to serve on",
					},
					&cli.StringFlag{
						Name:    "dir",
						Aliases: []string{"d"},
						Usage:   "Directory containing videos",
					},
				},
				Action: stream.ServeAction,
			},
			{
				Name:   "preprocess",
				Usage:  "Preprocess videos for HLS streaming",
				Action: stream.PreprocessAction,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "dir",
						Aliases:  []string{"d"},
						Usage:    "Directory containing videos",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "output",
						Aliases:  []string{"o"},
						Usage:    "Output directory for processed files",
						Required: true,
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
