// Package providers defines interfaces for AI providers.
package providers

import (
	"context"

	"github.com/rafa-mori/grompt/internal/types"
	"github.com/rafa-mori/logz"
)

type ProviderOpts[T any] = types.ProviderOpts[T]

type ProviderCtl[T any, C chan T] = types.ProviderCtl[T, C]

// Provider represents an AI provider interface
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

type Capabilities = types.Capabilities

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

// Initialize creates and returns all available providers
func Initialize[F types.APIConfig | types.OpenAIAPI | types.ClaudeAPI | types.GeminiAPI | types.DeepSeekAPI | types.OllamaAPI | types.ChatGPTAPI](
	bindAddr,
	port,
	openAIKey,
	deepSeekKey,
	ollamaEndpoint,
	claudeKey,
	geminiKey,
	chatGPTKey string,
	logger logz.Logger,
) []Provider[F] {

	if bindAddr == "" &&
		port == "" &&
		openAIKey == "" &&
		deepSeekKey == "" &&
		ollamaEndpoint == "" &&
		claudeKey == "" &&
		geminiKey == "" &&
		chatGPTKey == "" {
		return []Provider[F]{}
	}

	var cfg = types.NewConfig(
		bindAddr,
		"8080",
		openAIKey,
		deepSeekKey,
		ollamaEndpoint,
		claudeKey,
		geminiKey,
		chatGPTKey,
		nil,
	)

	cfg.Logger = logger

	var providers []Provider[F]
	if claudeKey != "" {
		apiCfg := types.NewClaudeAPI(claudeKey)
		prvdr := NewProvider("claude", claudeKey, "v1", apiCfg.(F), cfg)
		providers = append(providers, prvdr)
	}
	if openAIKey != "" {
		apiCfg := types.NewOpenAIAPI(openAIKey)
		providers = append(providers, NewProvider("openai", openAIKey, "v1", apiCfg.(F), cfg))
	}
	if deepSeekKey != "" {
		apiCfg := types.NewDeepSeekAPI(deepSeekKey)
		providers = append(providers, NewProvider("deepseek", deepSeekKey, "v1", apiCfg.(F), cfg))
	}
	if ollamaEndpoint != "" {
		apiCfg := types.NewOllamaAPI(ollamaEndpoint)
		providers = append(providers, NewProvider("ollama", ollamaEndpoint, "v1", apiCfg.(F), cfg))
	}
	if geminiKey != "" {
		apiCfg := types.NewGeminiAPI(geminiKey)
		providers = append(providers, NewProvider("gemini", geminiKey, "v1", apiCfg.(F), cfg))
	}
	if chatGPTKey != "" {
		apiCfg := types.NewChatGPTAPI(chatGPTKey)
		providers = append(providers, NewProvider("chatgpt", chatGPTKey, "v1", apiCfg.(F), cfg))
	}

	return providers
}
