// Package grompt provides an interface for modules that can be used with the grompt command-line tool.
// It also exposes prompt engineering capabilities for use as a library.
package grompt

import (
	"github.com/rafa-mori/grompt/factory/providers"
	"github.com/rafa-mori/grompt/internal/engine"
	m "github.com/rafa-mori/grompt/internal/module"
	sec "github.com/rafa-mori/grompt/internal/module/control"
	st "github.com/rafa-mori/grompt/internal/module/control"
	"github.com/rafa-mori/grompt/internal/types"
	"github.com/spf13/cobra"
)

// This file/package allows the grompt module to be used as a library.
// It defines the Grompt interface which can be implemented by any module
// that wants to be part of the grompt ecosystem.

// Grompt represents the main CLI interface
type Grompt interface {
	// Alias returns the alias for the command.
	Alias() string
	// ShortDescription returns a brief description of the command.
	ShortDescription() string
	// LongDescription returns a detailed description of the command.
	LongDescription() string
	// Usage returns the usage string for the command.
	Usage() string
	// Examples returns a list of example usages for the command.
	Examples() []string
	// Active returns true if the command is active and should be executed.
	Active() bool
	// Module returns the name of the module.
	Module() string
	// Execute runs the command and returns an error if it fails.
	Execute() error
	// Command returns the cobra.Command associated with this module.
	Command() *cobra.Command
}

// PromptEngine exposes the core prompt engineering functionality
type PromptEngine[F types.APIConfig | types.OpenAIAPI | types.ClaudeAPI | types.GeminiAPI | types.DeepSeekAPI | types.OllamaAPI | types.ChatGPTAPI] interface {
	// ProcessPrompt processes a prompt with variables and returns the result
	ProcessPrompt(template string, vars map[string]interface{}) (*Result, error)

	// GetProviders returns available AI providers
	GetProviders() []Provider[F]

	// GetHistory returns the prompt history
	GetHistory() []Result

	// SaveToHistory saves a prompt/response pair to history
	SaveToHistory(prompt, response string) error

	// BatchProcess processes multiple prompts concurrently
	BatchProcess(prompts []string, vars map[string]interface{}) ([]Result, error)
}

// NewGrompt creates a new Grompt CLI instance
func NewGrompt() Grompt {
	return m.RegX()
}

// NewPromptEngine creates a new prompt engineering engine for library use
func NewPromptEngine[F types.APIConfig | types.OpenAIAPI | types.ClaudeAPI | types.GeminiAPI | types.DeepSeekAPI | types.OllamaAPI | types.ChatGPTAPI](config Config) PromptEngine[F] {
	return engine.NewEngine[F](config)
}

// DefaultConfig returns a default configuration for the prompt engine
func DefaultConfig(configFilePath string) Config {
	return &types.Config{
		Port:           "8080",
		ClaudeAPIKey:   "",
		OpenAIAPIKey:   "",
		DeepSeekAPIKey: "",
		GeminiAPIKey:   "",
		ChatGPTAPIKey:  "",
		OllamaEndpoint: "http://localhost:11434",
		Debug:          false,
	}
}

// Exposed types for external use - avoid interface{} in consumer code

// Result exposes the engine.Result type
type Result = engine.Result

// Config exposes the types.Config interface
type Config = types.IConfig

// Provider exposes the providers.Provider interface
type Provider[F types.APIConfig | types.OpenAIAPI | types.ClaudeAPI | types.GeminiAPI | types.DeepSeekAPI | types.OllamaAPI | types.ChatGPTAPI] = providers.Provider[F]

type APIConfig = types.IAPIConfig

// --- Bitflags de Segurança ---
type SecFlag = sec.SecFlag

const (
	SecNone         = sec.SecNone
	SecAuth         = sec.SecAuth
	SecSanitize     = sec.SecSanitize
	SecSanitizeBody = sec.SecSanitizeBody
)

// --- Registrador Atômico de Flags ---
type FlagReg32A[T ~uint32] = sec.FlagReg32A[T]

// --- Job States/Flags ---
type JobFlag = st.JobFlag

const (
	JobPendingA         = st.JobPendingA
	JobRunningA         = st.JobRunningA
	JobCancelRequestedA = st.JobCancelRequestedA
	JobRetryingA        = st.JobRetryingA
	JobCompletedA       = st.JobCompletedA
	JobFailedA          = st.JobFailedA
	JobTimedOutA        = st.JobTimedOutA
)

type FlagReg32[T ~uint32] = st.FlagReg32[T]

var ErrTerminal = st.ErrTerminal
