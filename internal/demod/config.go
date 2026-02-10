package demod

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Version int      `toml:"version"`
	Modules []Module `toml:"modules"`
}

type Path struct {
	Src string `toml:"src"`
	As  string `toml:"as"`
}

type Module struct {
	Name     string `toml:"name"`
	Repo     string `toml:"repo"`
	Revision string `toml:"revision"`
	Dest     string `toml:"dest"`
	Paths    []Path `toml:"paths"`
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
		seen := make(map[string]struct{})
		for j, p := range mod.Paths {
			if p.Src == "" {
				return nil, fmt.Errorf("modules[%d] (%s): paths[%d].src is required", i, mod.Name, j)
			}

			destPath := p.As
			if destPath == "" {
				destPath = p.Src
			}

			cleaned := filepath.Clean(destPath)
			if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
				return nil, fmt.Errorf("modules[%d] (%s): paths[%d] has invalid dest path %q: path traversal is not allowed", i, mod.Name, j, destPath)
			}

			if _, ok := seen[cleaned]; ok {
				return nil, fmt.Errorf("modules[%d] (%s): duplicate dest path %q in paths", i, mod.Name, cleaned)
			}
			seen[cleaned] = struct{}{}
		}
	}

	return &cfg, nil
}
