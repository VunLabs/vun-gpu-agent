package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Config contains the settings required by the GPU agent.
type Config struct {
	ListenAddress string
	LogLevel      string
}

// Load reads a simple key/value YAML configuration file.
//
// The agent configuration intentionally supports only the scalar settings
// defined by Config, avoiding a runtime dependency for the initial scaffold.
func Load(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config: %w", err)
	}
	defer file.Close()

	cfg := Config{
		ListenAddress: ":8080",
		LogLevel:      "info",
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, ":")
		if !found {
			return Config{}, fmt.Errorf("invalid config line %q", line)
		}

		value = strings.Trim(strings.TrimSpace(value), `"'`)
		switch strings.TrimSpace(key) {
		case "listen_address":
			if value == "" {
				return Config{}, errors.New("listen_address cannot be empty")
			}
			cfg.ListenAddress = value
		case "log_level":
			if value == "" {
				return Config{}, errors.New("log_level cannot be empty")
			}
			cfg.LogLevel = value
		default:
			return Config{}, fmt.Errorf("unknown config key %q", strings.TrimSpace(key))
		}
	}
	if err := scanner.Err(); err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	return cfg, nil
}
