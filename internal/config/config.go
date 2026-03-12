package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	GeminiAPIKey string
	Model        string
	MaxTokens    int
}

func Load() (*Config, error) {
	key := os.Getenv("GEMINI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
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
