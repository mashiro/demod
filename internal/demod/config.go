package demod

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Version int      `toml:"version"`
	Modules []Module `toml:"modules"`
}

type Module struct {
	Name        string   `toml:"name"`
	Repo        string   `toml:"repo"`
	Revision    string   `toml:"revision"`
	Dest        string   `toml:"dest"`
	Paths       []string `toml:"paths"`
	StripPrefix string   `toml:"stripPrefix"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.Version != 1 {
		return nil, fmt.Errorf("unsupported config version: %d (expected 1)", cfg.Version)
	}

	for i, mod := range cfg.Modules {
		if mod.Name == "" {
			return nil, fmt.Errorf("modules[%d]: name is required", i)
		}
		if mod.Repo == "" {
			return nil, fmt.Errorf("modules[%d] (%s): repo is required", i, mod.Name)
		}
		if mod.Revision == "" {
			return nil, fmt.Errorf("modules[%d] (%s): revision is required", i, mod.Name)
		}
		if mod.Dest == "" {
			return nil, fmt.Errorf("modules[%d] (%s): dest is required", i, mod.Name)
		}
		if len(mod.Paths) == 0 {
			return nil, fmt.Errorf("modules[%d] (%s): paths is required", i, mod.Name)
		}
	}

	return &cfg, nil
}
