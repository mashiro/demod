package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mashiro/demod/internal/demod"
	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:  "demod",
		Usage: "Declarative git module synchronizer",
		Commands: []*cli.Command{
			{
				Name:  "sync",
				Usage: "Sync all modules defined in config",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "file",
						Aliases: []string{"f"},
						Value:   "demod.toml",
						Usage:   "Path to config file",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cfgPath := cmd.String("file")
					cfg, err := demod.Load(cfgPath)
					if err != nil {
						return err
					}
					return demod.SyncAll(cfg)
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
