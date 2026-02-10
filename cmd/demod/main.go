package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"github.com/mashiro/demod/internal/demod"
	"github.com/urfave/cli/v3"
)

var version = "dev"

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
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Value:   "text",
				Usage:   "Log format (text, json)",
			},
			&cli.BoolFlag{
				Name:  "no-color",
				Usage: "Disable colored output",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "version",
				Usage: "Print the version",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println(version)
					return nil
				},
			},
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
					logger := buildLogger(cmd.Root().String("format"), cmd.Root().Bool("no-color"))
					return demod.SyncAll(cfg, demod.SyncOptions{
						DryRun: cmd.Bool("dry-run"),
						Logger: logger,
					})
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func buildLogger(format string, noColor bool) *slog.Logger {
	switch format {
	case "json":
		return slog.New(slog.NewJSONHandler(os.Stderr, nil))
	default:
		return slog.New(demod.NewModuleHandler(
			tint.NewHandler(os.Stderr, &tint.Options{
				TimeFormat: " ",
				NoColor:    noColor,
			}),
		))
	}
}
