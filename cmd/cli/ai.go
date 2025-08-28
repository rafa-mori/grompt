package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/spf13/cobra"

	gl "github.com/rafa-mori/grompt/internal/module/logger"
	t "github.com/rafa-mori/grompt/internal/types"
	"github.com/rafa-mori/grompt/utils"
	l "github.com/rafa-mori/logz"
)

// getProviderAPIKey returns the API key only for the matching provider
func getProviderAPIKey(targetProvider, currentProvider, apiKey string) string {
	if currentProvider == targetProvider && apiKey != "" {
		return apiKey
	}
	return ""
}

// setupConfig creates configuration with proper API key distribution
func setupConfig(configFile, provider, apiKey, ollamaEndpoint string) (t.IConfig, error) {
	var cfg t.IConfig
	var err error

	if configFile != "" {
		cfg, err = loadConfigFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("error loading configuration file: %v", err)
		}
		gl.Log("info", "Configuration loaded from file.")
	} else {
		cfg = t.NewConfig(
			utils.GetEnvOr("BIND_ADDR", ""),
			utils.GetEnvOr("PORT", ""),
			utils.GetEnvOr("OPENAI_API_KEY", getProviderAPIKey("openai", provider, apiKey)),
			utils.GetEnvOr("DEEPSEEK_API_KEY", getProviderAPIKey("deepseek", provider, apiKey)),
			utils.GetEnvOr("OLLAMA_ENDPOINT", ollamaEndpoint),
			utils.GetEnvOr("CLAUDE_API_KEY", getProviderAPIKey("claude", provider, apiKey)),
			utils.GetEnvOr("GEMINI_API_KEY", getProviderAPIKey("gemini", provider, apiKey)),
			utils.GetEnvOr("CHATGPT_API_KEY", getProviderAPIKey("chatgpt", provider, apiKey)),
			gl.GetLogger[l.Logger](nil),
		)
	}

	if cfg == nil {
		return nil, fmt.Errorf("configuration not loaded")
	}

	return cfg, nil
}

// setupProvider initializes and validates the AI provider
func setupProvider(cfg t.IConfig, provider, apiKey string) (t.IAPIConfig, string, error) {
	if provider == "" {
		provider = getDefaultProvider(cfg)
	}

	// providerObj := t.NewProvider(
	// 	provider,
	// 	utils.GetEnvOr(fmt.Sprintf("%s_API_KEY", strings.ToUpper(provider)), ""),
	// 	cfg,
	// )

	apiCfg := cfg.GetAPIConfig(provider)

	apiC, ok := apiCfg.(*t.APIConfig)
	if !ok {
		return nil, "", fmt.Errorf("invalid API config for provider: %s", provider)
	}

	providerObj := t.NewProvider(provider, *apiC, cfg)
	if providerObj == nil {
		return nil, "", fmt.Errorf("unknown provider: %s", provider)
	}

	if providerObj.VConfig == nil {
		return nil, "", fmt.Errorf("provider '%s' is not configured", provider)
	}

	providerObj.VConfig.SetAPIKey(provider, apiKey)
	apiConfig := providerObj.VConfig.GetAPIConfig(providerObj.Name())
	if apiConfig == nil {
		return nil, "", fmt.Errorf("provider '%s' is not configured or available", provider)
	}

	if !providerObj.IsAvailable(context.Background()) {
		return nil, "", fmt.Errorf("provider '%s' is not available. Please check your API key and configuration", provider)
	}

	return apiConfig, provider, nil
}

// AICmdList returns all AI-related commands
func AICmdList() []*cobra.Command {
	return []*cobra.Command{
		askCommand(),
		generateCommand(),
		chatCommand(),
	}
}

// askCommand handles direct prompt requests to AI providers
func askCommand() *cobra.Command {
	var (
		debug      bool
		prompt     string
		provider   string
		model      string
		maxTokens  int
		configFile string
		// API Keys
		apiKey         string
		ollamaEndpoint string
	)

	cmd := &cobra.Command{
		Use:   "ask",
		Short: "Ask a direct question to an AI provider",
		Long: `Send a direct prompt to an AI provider without starting the server.

Examples:
  grompt ask --prompt "What is Go programming?" --provider gemini
  grompt ask --prompt "Explain REST APIs" --provider openai --model gpt-4
  grompt ask --prompt "Write a poem about code" --provider claude --max-tokens 500`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if debug {
				gl.GetLogger[l.Logger](nil)
				gl.SetDebug(true)
			}

			if len(prompt) == 0 {
				gl.Log("fatal", "Prompt cannot be empty. Use --prompt flag")
			}

			// Setup configuration
			cfg, err := setupConfig(configFile, provider, apiKey, ollamaEndpoint)
			if err != nil {
				gl.Log("fatal", err.Error())
			}

			// Setup provider
			apiConfig, provider, err := setupProvider(cfg, provider, apiKey)
			if err != nil {
				gl.Log("fatal", err.Error())
			}

			// Set default model if not specified
			if model == "" {
				models := apiConfig.GetCommonModels()
				if len(models) > 0 {
					model = models[0]
				}
			}

			// Set default max tokens
			if maxTokens <= 0 {
				maxTokens = 1000
			}

			gl.Log("info", fmt.Sprintf("🤖 Asking %s: %s", provider, truncateString(prompt, 60)))

			response, err := apiConfig.Complete(prompt, maxTokens, model)
			if err != nil {
				return fmt.Errorf("failed to get response from %s: %v", provider, err)
			}

			fmt.Printf("\n🎯 **%s Response (%s):**\n\n%s\n\n",
				strings.ToUpper(provider), model, response)

			return nil
		},
	}

	cmd.Flags().BoolVarP(&debug, "debug", "D", false, "Enable debug mode")
	cmd.Flags().StringVarP(&prompt, "prompt", "p", "", "The prompt to send to AI (required)")
	cmd.Flags().StringVarP(&provider, "provider", "P", "", "AI provider (openai, claude, gemini, deepseek, ollama)")
	cmd.Flags().StringVarP(&model, "model", "m", "", "Model to use (provider specific)")
	cmd.Flags().IntVarP(&maxTokens, "max-tokens", "t", 1000, "Maximum tokens in response")
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Config file path")

	// API Key flags
	cmd.Flags().StringVar(&apiKey, "apikey", "", "API key")
	cmd.Flags().StringVar(&ollamaEndpoint, "ollama-endpoint", "http://localhost:11434", "Ollama endpoint")

	cmd.MarkFlagRequired("prompt")

	return cmd
}

// generateCommand handles prompt engineering from ideas
func generateCommand() *cobra.Command {
	var (
		debug       bool
		ideas       []string
		purpose     string
		purposeType string
		lang        string
		maxTokens   int
		provider    string
		model       string
		configFile  string
		output      string
		// API Keys
		apiKey         string
		ollamaEndpoint string
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate professional prompts from raw ideas using prompt engineering",
		Long: `Transform raw, unorganized ideas into structured, professional prompts using AI-powered prompt engineering.

Examples:
  grompt generate --ideas "API design,REST,security" --purpose "Tutorial" --provider gemini
  grompt generate --ideas "machine learning,python,beginners" --purpose-type "Educational" --lang "english"
  grompt generate --ideas "docker,kubernetes,deployment" --output prompt.md --provider claude`,
		Run: func(cmd *cobra.Command, args []string) {
			if debug {
				gl.GetLogger[l.Logger](nil)
				gl.SetDebug(true)
			}

			if len(ideas) == 0 {
				gl.Log("fatal", "At least one idea is required. Use --ideas flag")
			}

			// Setup configuration
			cfg, err := setupConfig(configFile, provider, apiKey, ollamaEndpoint)
			if err != nil {
				gl.Log("fatal", err.Error())
			}

			// Setup provider
			apiConfig, provider, err := setupProvider(cfg, provider, apiKey)
			if err != nil {
				gl.Log("fatal", err.Error())
			}

			// Set defaults
			if lang == "" {
				lang = "english"
			}
			if maxTokens <= 0 {
				maxTokens = 2000
			}
			if purposeType == "" {
				purposeType = "code"
			}

			// Set default model if not specified
			if model == "" {
				models := apiConfig.GetCommonModels()
				if len(models) > 0 {
					model = models[0]
				}
			}

			gl.Log("info", fmt.Sprintf("🔨 Engineering prompt from %d ideas using %s", len(ideas), strings.ToTitleSpecial(unicode.CaseRanges, provider)))

			// Use the same prompt engineering logic from the server
			engineeringPrompt := cfg.GetBaseGenerationPrompt(ideas, purpose, purposeType, lang, maxTokens)

			response, err := apiConfig.Complete(engineeringPrompt, maxTokens, model)
			if err != nil {
				gl.Log("fatal", fmt.Sprintf("Error generating prompt: %v", err))
			}

			result := fmt.Sprintf("# Generated Prompt (%s - %s)\n\n%s", provider, model, response)

			// Output to file or stdout
			if output != "" {
				err := os.WriteFile(output, []byte(result), 0644)
				if err != nil {
					gl.Log("fatal", fmt.Sprintf("Error saving prompt to file: %v", err))
				}
				gl.Log("success", fmt.Sprintf("✅ Prompt saved to %s", output))
			} else {
				fmt.Println(result)
			}
		},
	}

	cmd.Flags().BoolVarP(&debug, "debug", "D", false, "Enable debug mode")
	cmd.Flags().StringSliceVarP(&ideas, "ideas", "i", []string{}, "Raw ideas (comma-separated or multiple flags)")
	cmd.Flags().StringVarP(&purpose, "purpose", "p", "", "Specific purpose description")
	cmd.Flags().StringVar(&purposeType, "purpose-type", "code", "Purpose type category")
	cmd.Flags().StringVarP(&lang, "lang", "l", "english", "Response language")
	cmd.Flags().IntVarP(&maxTokens, "max-tokens", "t", 2048, "Maximum tokens in response")
	cmd.Flags().StringVarP(&provider, "provider", "P", "", "AI provider")
	cmd.Flags().StringVarP(&model, "model", "m", "", "Model to use")
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Config file path")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file (default: stdout)")

	// API Key flags
	cmd.Flags().StringVar(&apiKey, "apikey", "", "API key")
	cmd.Flags().StringVar(&ollamaEndpoint, "ollama-endpoint", "http://localhost:11434", "Ollama endpoint")

	cmd.MarkFlagRequired("ideas")

	return cmd
}

// chatCommand provides interactive chat with AI providers
func chatCommand() *cobra.Command {
	var (
		debug      bool
		provider   string
		model      string
		maxTokens  int
		configFile string
		// API Keys
		apiKey         string
		ollamaEndpoint string
	)

	cmd := &cobra.Command{
		Use:   "chat",
		Short: "Start an interactive chat session with an AI provider",
		Long: `Start an interactive chat session where you can have a conversation with an AI provider.

Examples:
  grompt chat --provider gemini
  grompt chat --provider openai --model gpt-4
  grompt chat --provider claude --max-tokens 500`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if debug {
				gl.GetLogger[l.Logger](nil)
				gl.SetDebug(true)
			}
			var cfg t.IConfig
			var err error

			if configFile != "" {
				cfg, err = loadConfigFile(configFile)
				if err != nil {
					gl.Log("fatal", fmt.Sprintf("Error loading configuration file: %v", err))
				}
				gl.Log("info", "Configuration loaded from file.")
			} else {
				cfg = t.NewConfig(
					utils.GetEnvOr("BIND_ADDR", ""),
					utils.GetEnvOr("PORT", ""),
					utils.GetEnvOr("OPENAI_API_KEY", apiKey),
					utils.GetEnvOr("DEEPSEEK_API_KEY", apiKey),
					utils.GetEnvOr("OLLAMA_ENDPOINT", ollamaEndpoint),
					utils.GetEnvOr("CLAUDE_API_KEY", apiKey),
					utils.GetEnvOr("GEMINI_API_KEY", apiKey),
					utils.GetEnvOr("CHATGPT_API_KEY", apiKey),
					gl.GetLogger[l.Logger](nil),
				)
			}

			if cfg == nil {
				gl.Log("fatal", "Configuration not loaded")
			}

			if provider == "" {
				provider = getDefaultProvider(cfg)
			}

			apiConfig := cfg.GetAPIConfig(provider)
			if apiConfig == nil {
				gl.Log("fatal", fmt.Sprintf("Provider '%s' is not configured", provider))
			}

			apiC, ok := apiConfig.(*t.APIConfig)
			if !ok {
				return fmt.Errorf("invalid API config for provider: %s", provider)
			}

			providerObj := t.NewProvider(
				provider,
				*apiC,
				cfg,
			)
			if providerObj == nil {
				gl.Log("fatal", fmt.Sprintf("Unknown provider: %s", provider))
			} else {
				if providerObj.VConfig == nil {
					gl.Log("fatal", fmt.Sprintf("Provider '%s' is not configured", provider))
				}
			}

			providerObj.VConfig.SetAPIKey(provider, apiKey)
			apiConfig = providerObj.VConfig.GetAPIConfig(providerObj.Name())
			if apiConfig == nil {
				gl.Log("fatal", fmt.Sprintf("Provider '%s' is not configured or available", provider))
			}

			if !providerObj.IsAvailable(context.Background()) {
				gl.Log("fatal", fmt.Sprintf("Provider '%s' is not available. Please check your API key and configuration.", provider))
			}

			// Set default model if not specified
			if model == "" {
				models := apiConfig.GetCommonModels()
				if len(models) > 0 {
					model = models[0]
				}
			}

			// Set default max tokens
			if maxTokens <= 0 {
				maxTokens = 1000
			}

			gl.Log("info", fmt.Sprintf("🤖 Starting chat with %s (%s)\n", strings.ToUpper(provider), model))
			gl.Log("info", "───────────────────────────────────────────────────")
			gl.Log("info", "💡 Type 'exit', 'quit', or 'bye' to end the conversation")
			gl.Log("info", "💡 Your API key is used only for this session and not stored")
			gl.Log("info", "───────────────────────────────────────────────────")

			// Start interactive chat loop
			scanner := bufio.NewScanner(os.Stdin)

			// Simple interactive loop
			for {
				gl.Log("info", "🧑 You:")

				// Read user input with error handling
				if !scanner.Scan() {
					if err := scanner.Err(); err != nil {
						return fmt.Errorf("error reading input: %v", err)
					}
					break // EOF
				}
				input := strings.TrimSpace(scanner.Text())

				// Check for exit commands
				if input == "exit" || input == "quit" || input == "bye" {
					gl.Log("info", "👋 Goodbye!")
					break
				}

				if input == "" {
					continue
				}

				gl.Log("info", "🤖 AI:")

				response, err := apiConfig.Complete(input, maxTokens, model)
				if err != nil {
					gl.Log("error", fmt.Sprintf("error getting response from %s: %v", provider, err))
					continue
				}

				gl.Log("answer", response)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&debug, "debug", "D", false, "Enable debug mode")
	cmd.Flags().StringVarP(&provider, "provider", "P", "", "AI provider")
	cmd.Flags().StringVarP(&model, "model", "m", "", "Model to use")
	cmd.Flags().IntVarP(&maxTokens, "max-tokens", "t", 1000, "Maximum tokens per response")
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Config file path")

	// API Key flags
	cmd.Flags().StringVar(&apiKey, "apikey", "", "API key")
	cmd.Flags().StringVar(&ollamaEndpoint, "ollama-endpoint", "http://localhost:11434", "Ollama endpoint")

	return cmd
}

// getDefaultProvider returns the first available provider
func getDefaultProvider(cfg t.IConfig) string {
	providers := []string{"gemini", "claude", "openai", "deepseek", "ollama", "chatgpt"}

	for _, provider := range providers {
		if apiConfig := cfg.GetAPIConfig(provider); apiConfig != nil && apiConfig.IsAvailable() {
			return provider
		}
	}

	return ""
}

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
