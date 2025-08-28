// Package engine provides the core functionality for the factory engine.
package engine

import (
	"github.com/rafa-mori/grompt/internal/engine"
	"github.com/rafa-mori/grompt/internal/types"
)

type Engine[F types.APIConfig | types.OpenAIAPI | types.ClaudeAPI | types.GeminiAPI | types.DeepSeekAPI | types.OllamaAPI | types.ChatGPTAPI] = engine.IEngine[F]

func NewEngine[F types.APIConfig | types.OpenAIAPI | types.ClaudeAPI | types.GeminiAPI | types.DeepSeekAPI | types.OllamaAPI | types.ChatGPTAPI](config types.IConfig) engine.IEngine[F] {
	return engine.NewEngine[F](config)
}
