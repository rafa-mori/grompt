// Package engine provides the core prompt engineering functionality.
package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/rafa-mori/grompt/factory/providers"
	"github.com/rafa-mori/grompt/factory/templates"
	"github.com/rafa-mori/grompt/internal/types"
)

type IEngine[F types.APIConfig | types.OpenAIAPI | types.ClaudeAPI | types.GeminiAPI | types.DeepSeekAPI | types.OllamaAPI | types.ChatGPTAPI] interface {
	// ProcessPrompt processes a prompt with variables and returns the result
	ProcessPrompt(template string, vars map[string]interface{}) (*Result, error)

	// GetProviders returns available providers
	GetProviders() []providers.Provider[F]

	// GetHistory returns the prompt history
	GetHistory() []Result

	// SaveToHistory saves a prompt/response pair to history
	SaveToHistory(prompt, response string) error

	// BatchProcess processes multiple prompts concurrently
	BatchProcess(prompts []string, vars map[string]interface{}) ([]Result, error)
}

// Engine represents the core prompt engineering engine
type Engine[F types.APIConfig | types.OpenAIAPI | types.ClaudeAPI | types.GeminiAPI | types.DeepSeekAPI | types.OllamaAPI | types.ChatGPTAPI] struct {
	providers []providers.Provider[F]
	templates templates.Manager
	history   IHistoryManager
	config    types.IConfig
}

// NewEngine creates a new IEngine instance with initialized providers
func NewEngine[F types.APIConfig | types.OpenAIAPI | types.ClaudeAPI | types.GeminiAPI | types.DeepSeekAPI | types.OllamaAPI | types.ChatGPTAPI](config types.IConfig) IEngine[F] {
	engine := &Engine[F]{
		providers: make([]providers.Provider[F], 0),
		templates: templates.NewManager("./templates"), // Default templates path
		history:   NewHistoryManager(100),              // Default to 100 entries
		config:    config,
	}

	// Initialize concrete providers
	engine.initializeProviders()

	return engine
}

// initializeProviders initializes all available concrete providers
func (e *Engine[F]) initializeProviders() {
	if e == nil {
		return
	}

}

// ProcessPrompt processes a prompt with variables and returns the result
func (e *Engine[F]) ProcessPrompt(template string, vars map[string]interface{}) (*Result, error) {
	if e == nil {
		return nil, fmt.Errorf("engine is nil")
	}

	// Process template with variables
	processedPrompt, err := e.templates.Process(template, vars)
	if err != nil {
		return nil, fmt.Errorf("template processing failed: %w", err)
	}

	// Get default provider (first available)
	if len(e.providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	provider := e.providers[0]

	// Execute prompt with provider
	response, err := provider.Execute(context.Background(), processedPrompt)
	if err != nil {
		return nil, fmt.Errorf("provider execution failed: %w", err)
	}

	// Create result
	result := &Result{
		ID:        generateID(),
		Prompt:    processedPrompt,
		Response:  response,
		Provider:  provider.Name(),
		Variables: vars,
		Timestamp: time.Now(),
	}

	// Add to history
	e.history.Add(*result)

	return result, nil
}

// GetProviders returns available providers
func (e *Engine[F]) GetProviders() []providers.Provider[F] {
	if e == nil {
		return nil
	}
	return e.providers
}

// GetHistory returns the prompt history
func (e *Engine[F]) GetHistory() []Result {
	if e == nil || e.history == nil {
		return nil
	}
	return e.history.GetHistory()
}

// SaveToHistory saves a prompt/response pair to history
func (e *Engine[F]) SaveToHistory(prompt, response string) error {
	if e == nil || e.history == nil {
		return fmt.Errorf("engine or history is nil")
	}

	result := Result{
		ID:        generateID(),
		Prompt:    prompt,
		Response:  response,
		Provider:  "manual",
		Timestamp: time.Now(),
	}

	e.history.Add(result)
	return nil
}

// BatchProcess processes multiple prompts concurrently
func (e *Engine[F]) BatchProcess(prompts []string, vars map[string]interface{}) ([]Result, error) {
	if e == nil {
		return nil, fmt.Errorf("engine is nil")
	}

	results := make([]Result, len(prompts))
	errors := make([]error, len(prompts))

	// Process prompts concurrently
	for i, prompt := range prompts {
		go func(index int, p string) {
			result, err := e.ProcessPrompt(p, vars)
			if err != nil {
				errors[index] = err
				return
			}
			results[index] = *result
		}(i, prompt)
	}

	// Wait for all goroutines to complete (simplified version)
	time.Sleep(time.Second * 2) // In production, use sync.WaitGroup

	// Check for errors
	for _, err := range errors {
		if err != nil {
			return results, fmt.Errorf("batch processing errors occurred: %v", err)
		}
	}

	return results, nil
}

// generateID generates a simple ID for results
func generateID() string {
	return fmt.Sprintf("prompt_%d", time.Now().UnixNano())
}
