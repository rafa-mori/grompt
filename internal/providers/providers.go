// Package providers provides concrete implementations of AI providers.
package providers

import (
	"context"

	"github.com/rafa-mori/grompt/internal/types"
)

type ProviderCtl[T any, C chan T] = types.ProviderCtl[T, C]
type ProviderOpts[T any] = types.ProviderOpts[T]

// Provider defines the interface for AI providers.
type Provider[F types.APIConfig | types.OpenAIAPI | types.ClaudeAPI | types.GeminiAPI | types.DeepSeekAPI | types.OllamaAPI | types.ChatGPTAPI] interface {
	Name() string
	Version() string
	GetAPI() *F

	// IsAvailable checks if the provider is configured and ready
	IsAvailable(ctx context.Context) bool
	// GetCapabilities returns provider-specific capabilities
	GetCapabilities(ctx context.Context) *types.Capabilities

	// Execute a prompt and return the response
	Execute(ctx context.Context, prompt string, opts ...any) (string, error)
	// Stream opens a streaming response for a prompt
	Stream(ctx context.Context, prompt string, opts ...any) (string, error)

	// Initialize the provider
	Initialize(ctx context.Context) error
	// Initialize the provider with options
	InitializeWithOptions(opts *ProviderOpts[F]) error
	// Get the provider control structure
	Control(ctx context.Context) *ProviderCtl[F, chan F]
	// Stop the provider
	Stop(ctx context.Context) error
}

// Individual provider constructors for engine initialization

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider[F types.APIConfig | types.OpenAIAPI | types.ClaudeAPI | types.GeminiAPI | types.DeepSeekAPI | types.OllamaAPI | types.ChatGPTAPI](apiCfg F, mainCfg types.IConfig) Provider[F] {
	if cfg, ok := mainCfg.(*types.Config); !ok {
		return nil
	} else {
		return &types.ProviderImpl[F]{
			VName:   "openai",
			VAPI:    apiCfg,
			VConfig: cfg,
		}
	}
}

// NewClaudeProvider creates a new Claude provider
func NewClaudeProvider[F types.APIConfig | types.OpenAIAPI | types.ClaudeAPI | types.GeminiAPI | types.DeepSeekAPI | types.OllamaAPI | types.ChatGPTAPI](apiCfg F, mainCfg types.IConfig) Provider[F] {
	if cfg, ok := mainCfg.(*types.Config); !ok {
		return nil
	} else {
		return &types.ProviderImpl[F]{
			VName:   "claude",
			VAPI:    apiCfg,
			VConfig: cfg,
		}
	}
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider[F types.APIConfig | types.OpenAIAPI | types.ClaudeAPI | types.GeminiAPI | types.DeepSeekAPI | types.OllamaAPI | types.ChatGPTAPI](apiCfg F, mainCfg types.IConfig) Provider[F] {
	if cfg, ok := mainCfg.(*types.Config); !ok {
		return nil
	} else {
		return &types.ProviderImpl[F]{
			VName:   "gemini",
			VAPI:    apiCfg,
			VConfig: cfg,
		}
	}
}

// NewDeepSeekProvider creates a new DeepSeek provider
func NewDeepSeekProvider[F types.APIConfig | types.OpenAIAPI | types.ClaudeAPI | types.GeminiAPI | types.DeepSeekAPI | types.OllamaAPI | types.ChatGPTAPI](apiCfg F, mainCfg types.IConfig) Provider[F] {
	if cfg, ok := mainCfg.(*types.Config); !ok {
		return nil
	} else {
		return &types.ProviderImpl[F]{
			VName:   "deepseek",
			VAPI:    apiCfg,
			VConfig: cfg,
		}
	}
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider[F types.APIConfig | types.OpenAIAPI | types.ClaudeAPI | types.GeminiAPI | types.DeepSeekAPI | types.OllamaAPI | types.ChatGPTAPI](apiCfg F, mainCfg types.IConfig) Provider[F] {
	if cfg, ok := mainCfg.(*types.Config); !ok {
		return nil
	} else {
		return &types.ProviderImpl[F]{
			VName:   "ollama",
			VAPI:    apiCfg,
			VConfig: cfg,
		}
	}
}

func NewChatGPTProvider[F types.APIConfig | types.OpenAIAPI | types.ClaudeAPI | types.GeminiAPI | types.DeepSeekAPI | types.OllamaAPI | types.ChatGPTAPI](apiCfg F, mainCfg types.IConfig) Provider[F] {
	if cfg, ok := mainCfg.(*types.Config); !ok {
		return nil
	} else {
		return &types.ProviderImpl[F]{
			VName:   "chatgpt",
			VAPI:    apiCfg,
			VConfig: cfg,
		}
	}
}

func NewProvider[F types.APIConfig | types.OpenAIAPI | types.ClaudeAPI | types.GeminiAPI | types.DeepSeekAPI | types.OllamaAPI | types.ChatGPTAPI](name, apiKey, version string, apiCfg F, mainCfg types.IConfig) Provider[F] {
	// Check if mainCfg is of the correct type
	if cfg, ok := mainCfg.(*types.Config); !ok {
		return nil
	} else {
		return &types.ProviderImpl[F]{
			VName:    name,
			VVersion: version,
			VAPI:     apiCfg,
			VConfig:  cfg,
		}
	}
}
