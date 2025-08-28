// Package config provides configuration management for the factory.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	providersPkg "github.com/rafa-mori/grompt/internal/providers"
	"github.com/rafa-mori/grompt/internal/types"
	"gopkg.in/yaml.v3"

	l "github.com/rafa-mori/logz"
)

type Config = types.IConfig

func NewConfig(
	bindAddr,
	port,
	openAIKey,
	deepSeekKey,
	ollamaEndpoint,
	claudeKey,
	geminiKey,
	chatGPTKey string,
	logger l.Logger,
) types.IConfig {
	return types.NewConfig(
		bindAddr,
		port,
		openAIKey,
		deepSeekKey,
		ollamaEndpoint,
		claudeKey,
		geminiKey,
		chatGPTKey,
		logger,
	)
}

func NewConfigFromFile(filePath string) types.IConfig {
	var cfg types.Config
	if _, statErr := os.Stat(filePath); statErr != nil {
		return &types.Config{}
	}
	switch fileExt := filepath.Ext(filePath); fileExt {
	case ".json":
		if err := readJSONFile(filePath, &cfg); err != nil {
			return &types.Config{}
		}
	case ".yaml", ".yml":
		if err := readYAMLFile(filePath, &cfg); err != nil {
			return &types.Config{}
		}
	default:
		return &types.Config{}
	}
	return &cfg
}

func readJSONFile(filePath string, cfg *types.Config) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(cfg)
}

func readYAMLFile(filePath string, cfg *types.Config) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	return decoder.Decode(cfg)
}

func NewProvider[F types.APIConfig | types.OpenAIAPI | types.ClaudeAPI | types.GeminiAPI | types.DeepSeekAPI | types.OllamaAPI | types.ChatGPTAPI](name, apiKey, version string) providersPkg.Provider[F] {
	cfg := types.NewConfig("", "", "", "", "", "", "", "", nil)
	// Initialize provider-specific configuration
	var providerConfig F
	switch name {
	case "openai":
		providerConfig = types.NewOpenAIAPI(apiKey).(F)
	case "claude":
		providerConfig = types.NewClaudeAPI(apiKey).(F)
	case "gemini":
		providerConfig = types.NewGeminiAPI(apiKey).(F)
	case "deepseek":
		providerConfig = types.NewDeepSeekAPI(apiKey).(F)
	case "ollama":
		providerConfig = types.NewOllamaAPI(apiKey).(F)
	case "chatgpt":
		providerConfig = types.NewChatGPTAPI(apiKey).(F)
	default:
		return nil
	}
	return providersPkg.NewProvider(name, apiKey, version, providerConfig, cfg)
}
