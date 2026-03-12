package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	GeminiAPIKey string
	Model        string
	MaxTokens    int
}

type configFile struct {
	GeminiAPIKey string `json:"gemini_api_key"`
}

func loadConfigFile() (*configFile, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(home, ".config", "opencraft", "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cf configFile
	if err := json.Unmarshal(data, &cf); err != nil {
		return nil, err
	}
	return &cf, nil
}

func Load() (*Config, error) {
	key := os.Getenv("GEMINI_API_KEY")
	if key == "" {
		if cf, err := loadConfigFile(); err == nil && cf.GeminiAPIKey != "" {
			key = cf.GeminiAPIKey
		}
	}

	if key == "" {
		return nil, fmt.Errorf(`opencraft: GEMINI_API_KEY not set

Set it via environment variable:
  export GEMINI_API_KEY=your-key

Or create a config file at ~/.config/opencraft/config.json:
  {"gemini_api_key": "your-key"}`)
	}

	model := os.Getenv("OPENCRAFT_MODEL")
	if model == "" {
		model = "gemini-3.1-flash-lite-preview"
	}

	maxTokens := 8096

	if s := os.Getenv("OPENCRAFT_MAX_TOKENS"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			maxTokens = n
		}
	}

	return &Config{
		GeminiAPIKey: key,
		Model:        model,
		MaxTokens:    maxTokens,
	}, nil
}
