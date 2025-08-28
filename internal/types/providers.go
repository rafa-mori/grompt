package types

import (
	"context"
	"fmt"
)

type ProviderOpts[T any] struct {
	*Mutexes

	Settings map[string]any
	History  []any
	Data     *T
}

func NewProviderOpts[T any]() *ProviderOpts[T] {
	return &ProviderOpts[T]{
		Mutexes:  NewMutexesType(),
		Settings: make(map[string]any),
		History:  make([]any, 0),
		Data:     new(T),
	}
}

type ProviderCtl[T any, C chan T] struct {
	*Mutexes
	*ProviderOpts[T]

	Ch C
}

// ProviderImpl wraps the types.IAPIConfig to implement providers.Provider
type ProviderImpl[F APIConfig | OpenAIAPI | ClaudeAPI | GeminiAPI | DeepSeekAPI | OllamaAPI | ChatGPTAPI] struct {
	VName    string
	VVersion string
	VAPI     F
	VConfig  *Config
}

// Capabilities describes what a provider can do
type Capabilities struct {
	MaxTokens         int      `json:"max_tokens"`
	SupportsBatch     bool     `json:"supports_batch"`
	SupportsStreaming bool     `json:"supports_streaming"`
	Models            []string `json:"models"`
	Pricing           *Pricing `json:"pricing,omitempty"`
}

// Pricing information for the provider
type Pricing struct {
	InputCostPer1K  float64 `json:"input_cost_per_1k"`
	OutputCostPer1K float64 `json:"output_cost_per_1k"`
	Currency        string  `json:"currency"`
}

// Name returns the provider name
func (cp *ProviderImpl[F]) Name() string {
	return cp.VName
}

// Version returns the provider version
func (cp *ProviderImpl[F]) Version() string {
	return cp.VVersion
}

// Execute sends a prompt to the provider and returns the response
func (cp *ProviderImpl[F]) Execute(ctx context.Context, prompt string, opts ...any) (string, error) {
	if cp == nil {
		return "", fmt.Errorf("provider is not available")
	}
	return any(cp.VAPI).(IAPIConfig).Complete(prompt, 2048, "")
}

// IsAvailable checks if the provider is configured and ready
func (cp *ProviderImpl[F]) IsAvailable(ctx context.Context) bool {
	if cp == nil {
		return false
	}
	return any(cp.VAPI).(IAPIConfig).IsAvailable()
}

// GetCapabilities returns provider-specific capabilities
func (cp *ProviderImpl[F]) GetCapabilities(ctx context.Context) *Capabilities {
	if cp == nil {
		return nil
	}
	var api IAPIConfig
	if cp.VConfig != nil {
		cfg := *cp.VConfig // Dereference pointer to get back real pointer access
		switch cp.VName {
		case "openai":
			api = cfg.GetAPIConfig("openai")
		case "claude":
			api = cfg.GetAPIConfig("claude")
		case "gemini":
			api = cfg.GetAPIConfig("gemini")
		case "deepseek":
			api = cfg.GetAPIConfig("deepseek")
		case "ollama":
			api = cfg.GetAPIConfig("ollama")
		default:
			return nil // No API config available for this provider
		}
	}
	if api == nil {
		return nil // No API config available for this provider
	}
	models, err := api.ListModels()
	if err != nil {
		return nil
	}
	return &Capabilities{
		MaxTokens:         getMaxTokensForProvider(cp.VName),
		SupportsBatch:     true,
		SupportsStreaming: false, // For now, streaming is not implemented
		Models:            models,
		Pricing:           getPricingForProvider(cp.VName),
	}
}

func (cp *ProviderImpl[F]) GetAPI() *F {
	if cp == nil {
		return nil
	}
	return &cp.VAPI
}

func (cp *ProviderImpl[F]) Stream(ctx context.Context, prompt string, opts ...any) (string, error) {
	if cp == nil {
		return "", fmt.Errorf("provider is not available")
	}
	api := cp.VAPI
	// TODO: Fix this shit
	return any(api).(IAPIConfig).StartStream(prompt, 2048, "")
}

func (cp *ProviderImpl[F]) Initialize(ctx context.Context) error {
	return nil
}

func (cp *ProviderImpl[F]) InitializeWithOptions(opts *ProviderOpts[F]) error {
	return nil
}

func (cp *ProviderImpl[F]) Control(ctx context.Context) *ProviderCtl[F, chan F] {
	return &ProviderCtl[F, chan F]{
		Mutexes:      NewMutexesType(),
		ProviderOpts: NewProviderOpts[F](),
		Ch:           make(chan F),
	}
}

func (cp *ProviderImpl[F]) Stop(ctx context.Context) error {
	return nil
}

// getMaxTokensForProvider returns max tokens for each provider
func getMaxTokensForProvider(providerName string) int {
	switch providerName {
	case "openai":
		return 4096
	case "claude":
		return 8192
	case "gemini":
		return 8192
	case "deepseek":
		return 4096
	case "ollama":
		return 2048
	default:
		return 2048
	}
}

// getPricingForProvider returns pricing information for each provider
func getPricingForProvider(providerName string) *Pricing {
	switch providerName {
	case "openai":
		return &Pricing{
			InputCostPer1K:  0.0015,
			OutputCostPer1K: 0.002,
			Currency:        "USD",
		}
	case "claude":
		return &Pricing{
			InputCostPer1K:  0.003,
			OutputCostPer1K: 0.015,
			Currency:        "USD",
		}
	case "gemini":
		return &Pricing{
			InputCostPer1K:  0.000125,
			OutputCostPer1K: 0.000375,
			Currency:        "USD",
		}
	case "deepseek":
		return &Pricing{
			InputCostPer1K:  0.00014,
			OutputCostPer1K: 0.00028,
			Currency:        "USD",
		}
	case "ollama":
		return &Pricing{
			InputCostPer1K:  0.0,
			OutputCostPer1K: 0.0,
			Currency:        "USD", // Free local model
		}
	default:
		return nil
	}
}

// NewProviders creates all available providers based on configuration
func NewProviders[F APIConfig | OpenAIAPI | ClaudeAPI | GeminiAPI | DeepSeekAPI | OllamaAPI | ChatGPTAPI](apiCfg F, config IConfig) []*ProviderImpl[F] {
	var activeProviders []*ProviderImpl[F]

	// List of all supported providers
	providerConfigs := []struct {
		name string
		key  string
	}{
		{"openai", "openai"},
		{"claude", "claude"},
		{"gemini", "gemini"},
		{"deepseek", "deepseek"},
		{"ollama", "ollama"},
		{"chatgpt", "chatgpt"},
	}
	for _, cfg := range providerConfigs {
		apiKey := config.GetAPIKey(cfg.key)
		if apiKey != "" || cfg.name == "ollama" { // Ollama can work without API key (local)
			var provider *ProviderImpl[F]
			switch cfg.name {
			case "openai":
				provider = NewOpenAIProvider(apiCfg, config)
			case "claude":
				provider = NewClaudeProvider(apiCfg, config)
			case "gemini":
				provider = NewGeminiProvider(apiCfg, config)
			case "deepseek":
				provider = NewDeepSeekProvider(apiCfg, config)
			case "ollama":
				provider = NewOllamaProvider(apiCfg, config)
			case "chatgpt":
				provider = NewChatGPTProvider(apiCfg, config)
			}
			if provider != nil {
				activeProviders = append(activeProviders, provider)
			}
		}
	}

	return activeProviders
}

// Individual provider constructors for engine initialization

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider[F APIConfig | OpenAIAPI | ClaudeAPI | GeminiAPI | DeepSeekAPI | OllamaAPI | ChatGPTAPI](apiCfg F, mainCfg IConfig) *ProviderImpl[F] {
	if cfg, ok := mainCfg.(*Config); !ok {
		return &ProviderImpl[F]{
			VName:   "openai",
			VAPI:    apiCfg,
			VConfig: cfg,
		}
	} else {
		return &ProviderImpl[F]{
			VName:   "openai",
			VAPI:    apiCfg,
			VConfig: cfg,
		}
	}
}

// NewClaudeProvider creates a new Claude provider
func NewClaudeProvider[F APIConfig | OpenAIAPI | ClaudeAPI | GeminiAPI | DeepSeekAPI | OllamaAPI | ChatGPTAPI](apiCfg F, mainCfg IConfig) *ProviderImpl[F] {
	if cfg, ok := mainCfg.(*Config); !ok {
		return nil
	} else {
		return &ProviderImpl[F]{
			VName:   "claude",
			VAPI:    apiCfg,
			VConfig: cfg,
		}
	}
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider[F APIConfig | OpenAIAPI | ClaudeAPI | GeminiAPI | DeepSeekAPI | OllamaAPI | ChatGPTAPI](apiCfg F, mainCfg IConfig) *ProviderImpl[F] {
	if cfg, ok := mainCfg.(*Config); !ok {
		return nil
	} else {
		return &ProviderImpl[F]{
			VName:   "gemini",
			VAPI:    apiCfg,
			VConfig: cfg,
		}
	}
}

// NewDeepSeekProvider creates a new DeepSeek provider
func NewDeepSeekProvider[F APIConfig | OpenAIAPI | ClaudeAPI | GeminiAPI | DeepSeekAPI | OllamaAPI | ChatGPTAPI](apiCfg F, mainCfg IConfig) *ProviderImpl[F] {
	if cfg, ok := mainCfg.(*Config); !ok {
		return nil
	} else {
		return &ProviderImpl[F]{
			VName:   "deepseek",
			VAPI:    apiCfg,
			VConfig: cfg,
		}
	}
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider[F APIConfig | OpenAIAPI | ClaudeAPI | GeminiAPI | DeepSeekAPI | OllamaAPI | ChatGPTAPI](apiCfg F, mainCfg IConfig) *ProviderImpl[F] {
	if cfg, ok := mainCfg.(*Config); !ok {
		return nil
	} else {
		return &ProviderImpl[F]{
			VName:   "ollama",
			VAPI:    apiCfg,
			VConfig: cfg,
		}
	}
}

func NewChatGPTProvider[F APIConfig | OpenAIAPI | ClaudeAPI | GeminiAPI | DeepSeekAPI | OllamaAPI | ChatGPTAPI](apiCfg F, mainCfg IConfig) *ProviderImpl[F] {
	if cfg, ok := mainCfg.(*Config); !ok {
		return nil
	} else {
		return &ProviderImpl[F]{
			VName:   "chatgpt",
			VAPI:    apiCfg,
			VConfig: cfg,
		}
	}
}

func NewProvider[F APIConfig | OpenAIAPI | ClaudeAPI | GeminiAPI | DeepSeekAPI | OllamaAPI | ChatGPTAPI](name string, apiCfg F, mainCfg IConfig) *ProviderImpl[F] {
	if cfg, ok := mainCfg.(*Config); !ok {
		return nil
	} else {
		return &ProviderImpl[F]{
			VName:   name,
			VAPI:    apiCfg,
			VConfig: cfg,
		}
	}
}
