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
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "demod.toml",
				Usage:   "Path to config file",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "sync",
				Usage: "Sync all modules defined in config",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "dry-run",
						Usage: "Show what would be synced without making changes",
					},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					cfgPath := cmd.Root().String("config")
					cfg, err := demod.Load(cfgPath)
					if err != nil {
						return err
					}
					return demod.SyncAll(cfg, cmd.Bool("dry-run"))
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
