// Package types defines the configuration and versioning for the Grompt application.
package types

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"time"

	gl "github.com/rafa-mori/grompt/internal/module/logger"
	vs "github.com/rafa-mori/grompt/internal/module/version"
	"github.com/rafa-mori/grompt/utils"
	l "github.com/rafa-mori/logz"
)

var (
	CurrentVersion string    = vs.GetVersion()
	LatestVersion  string    = vs.GetLatestVersionFromGit()
	LastCheckTime  time.Time = time.Now()
)

func init() {
	// Initialize the CurrentVersion and LatestVersion
	// This will run in a goroutine to avoid blocking the main execution
	// and will check for the latest version from Git if needed.
	// If the CurrentVersion is not set, it will use the version from the version package
	// and will update the LatestVersion if the last check was more than 24 hours ago
	// or if it is the first run.
	ctx := context.Background()
	cancel, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()

	// Use a goroutine to avoid blocking the main execution
	go func(cancel context.Context) {
		select {
		case <-cancel.Done():
			return // Exit if the context is cancelled
		default:
			if CurrentVersion == "" {
				CurrentVersion = vs.GetVersion()
			}
			if LastCheckTime.IsZero() || LastCheckTime.Before(time.Now().Add(-24*time.Hour)) {
				// Check for the latest version from Git
				LatestVersion = vs.GetLatestVersionFromGit()
				if LatestVersion == "" {
					LatestVersion = CurrentVersion // Fallback to current version if check fails
				}
				LastCheckTime = time.Now()
			}
		}
	}(cancel)

	// // Ensure that the CurrentVersion is always set
	// if CurrentVersion == "" {
	// 	CurrentVersion = vs.GetVersion()
	// }
	// if LatestVersion == "" {
	// 	LatestVersion = vs.GetLatestVersionFromGit()
	// }
}

var AppVersion = vs.GetVersion()

const (
	AppName     = "Grompt"
	DefaultPort = "8080"
)

type IAPIConfig interface {
	IsAvailable() bool
	IsDemoMode() bool
	GetVersion() string
	ListModels() ([]string, error)
	GetCommonModels() []string

	Complete(prompt string, maxTokens int, model string) (string, error)

	StartStream(prompt string, maxTokens int, model string) (string, error)
	StopStream() error
}

type APIConfig struct {
	apiKey     string
	baseURL    string
	version    string
	httpClient *http.Client
	demoMode   bool
}

func (c *APIConfig) IsAvailable() bool {
	return c != nil && c.apiKey != ""
}
func (c *APIConfig) IsDemoMode() bool {
	return c != nil && c.demoMode
}
func (c *APIConfig) GetVersion() string {
	return c.version
}
func (c *APIConfig) ListModels() ([]string, error) {
	if c == nil {
		return nil, fmt.Errorf("Config is nil")
	}
	return nil, nil
}
func (c *APIConfig) GetCommonModels() []string {
	if c == nil {
		return nil
	}
	return nil
}
func (c *APIConfig) Complete(prompt string, maxTokens int, model string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("Config is nil")
	}
	return "", nil
}
func (c *APIConfig) StartStream(prompt string, maxTokens int, model string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("Config is nil")
	}
	return "", nil
}
func (c *APIConfig) StopStream() error {
	if c == nil {
		return fmt.Errorf("Config is nil")
	}
	return nil
}

type IConfig interface {
	GetAPIConfig(provider string) IAPIConfig
	GetPort() string
	GetAPIKey(provider string) string
	SetAPIKey(provider string, key string) error
	GetAPIEndpoint(provider string) string
	GetBaseGenerationPrompt(ideas []string, purpose, purposeType, lang string, maxLength int) string
}

type Config struct {
	Logger         l.Logger
	BindAddr       string `json:"bind_addr,omitempty" gorm:"default:'localhost'"`
	Port           string `json:"port" gorm:"default:8080"`
	OpenAIAPIKey   string `json:"openai_api_key,omitempty" gorm:"default:''"`
	DeepSeekAPIKey string `json:"deepseek_api_key,omitempty" gorm:"default:''"`
	OllamaEndpoint string `json:"ollama_endpoint,omitempty" gorm:"default:'http://localhost:11434'"`
	ClaudeAPIKey   string `json:"claude_api_key,omitempty" gorm:"default:''"`
	GeminiAPIKey   string `json:"gemini_api_key,omitempty" gorm:"default:''"`
	ChatGPTAPIKey  string `json:"chatgpt_api_key,omitempty" gorm:"default:''"`
	Debug          bool   `json:"debug" gorm:"default:false"`
}

func NewConfig(bindAddr, port, openAIKey, deepSeekKey, ollamaEndpoint, claudeKey, geminiKey, chatGPTKey string, logger l.Logger) *Config {
	if logger == nil {
		logger = gl.GetLogger[l.Logger](nil)
	}
	return &Config{
		Logger:         logger,
		BindAddr:       bindAddr,
		Port:           port,
		OpenAIAPIKey:   openAIKey,
		DeepSeekAPIKey: deepSeekKey,
		OllamaEndpoint: ollamaEndpoint,
		ClaudeAPIKey:   claudeKey,
		GeminiAPIKey:   geminiKey,
	}
}

func (c *Config) GetAPIConfig(provider string) IAPIConfig {
	if c == nil {
		gl.Log("error", "Config is nil")
		return nil
	}
	switch provider {
	case "openai":
		return NewOpenAIAPI(c.GetAPIKey("openai"))
	case "deepseek":
		return NewDeepSeekAPI(c.GetAPIKey("deepseek"))
	case "ollama":
		return NewOllamaAPI(c.GetAPIEndpoint("ollama"))
	case "claude":
		return NewClaudeAPI(c.GetAPIKey("claude"))
	case "gemini":
		return NewGeminiAPI(c.GetAPIKey("gemini"))
	case "chatgpt":
		return NewChatGPTAPI(c.GetAPIKey("chatgpt"))
	default:
		return nil
	}
}

func (c *Config) GetPort() string {
	if c.Port == "" {
		return DefaultPort
	}
	return c.Port
}

func (c *Config) GetAPIKey(provider string) string {
	switch provider {
	case "openai":
		if c.OpenAIAPIKey == "" {
			c.OpenAIAPIKey = utils.GetEnvOr("OPENAI_API_KEY", c.OpenAIAPIKey)
		}
		return c.OpenAIAPIKey
	case "deepseek":
		if c.DeepSeekAPIKey == "" {
			c.DeepSeekAPIKey = utils.GetEnvOr("DEEPSEEK_API_KEY", c.DeepSeekAPIKey)
		}
		return c.DeepSeekAPIKey
	case "ollama":
		if c.OllamaEndpoint == "" {
			c.OllamaEndpoint = utils.GetEnvOr("OLLAMA_ENDPOINT", c.OllamaEndpoint)
		}
		return c.OllamaEndpoint
	case "claude":
		if c.ClaudeAPIKey == "" {
			c.ClaudeAPIKey = utils.GetEnvOr("CLAUDE_API_KEY", c.ClaudeAPIKey)
		}
		return c.ClaudeAPIKey
	case "gemini":
		if c.GeminiAPIKey == "" {
			c.GeminiAPIKey = utils.GetEnvOr("GEMINI_API_KEY", c.GeminiAPIKey)
		}
		return c.GeminiAPIKey
	case "chatgpt":
		if c.ChatGPTAPIKey == "" {
			c.ChatGPTAPIKey = utils.GetEnvOr("CHATGPT_API_KEY", c.ChatGPTAPIKey)
		}
		return c.ChatGPTAPIKey
	default:
		return ""
	}
}

func (c *Config) SetAPIKey(provider string, key string) error {
	switch provider {
	case "openai":
		c.OpenAIAPIKey = key
	case "deepseek":
		c.DeepSeekAPIKey = key
	case "ollama":
		c.OllamaEndpoint = key
	case "claude":
		c.ClaudeAPIKey = key
	case "gemini":
		c.GeminiAPIKey = key
	case "chatgpt":
		c.ChatGPTAPIKey = key
	default:
		return fmt.Errorf("unknown provider: %s", provider)
	}
	return nil
}

func (c *Config) GetAPIEndpoint(provider string) string {
	if provider == "ollama" {
		return c.OllamaEndpoint
	}
	return ""
}

func (c *Config) StartStream(prompt string, maxTokens int, model string) (string, error) {
	if c == nil {
		return "", fmt.Errorf("Config is nil")
	}
	api := c.GetAPIConfig(model)
	if api == nil {
		return "", fmt.Errorf("API config is nil for model: %s", model)
	}
	return api.StartStream(prompt, maxTokens, "")
}

func (c *Config) StopStream() error {
	if c == nil {
		return fmt.Errorf("Config is nil")
	}
	api := c.GetAPIConfig("ollama")
	if api == nil {
		return fmt.Errorf("API config is nil for model: %s", "ollama")
	}
	return api.StopStream()
}

func (c *Config) checkOllamaConnection() bool {
	ip, err := netip.ParseAddr(c.OllamaEndpoint)
	if err != nil {
		return false
	}
	conn, err := net.DialTimeout("tcp", ip.String()+":11434", 2*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// GetBaseGenerationPrompt transforms raw ideas into a structured prompt engineering template.
// This method applies professional prompt engineering techniques to convert unorganized ideas
// into well-structured, effective prompts for AI interactions.
func (c *Config) GetBaseGenerationPrompt(ideas []string, purpose, purposeType, lang string, maxLength int) string {
	// Set default values
	if ideas == nil {
		ideas = []string{}
	}
	if lang == "" {
		lang = "english"
	}
	if maxLength <= 0 {
		maxLength = 2000
	}
	if purposeType == "" {
		purposeType = "Code"
	}

	// Build ideas list
	ideasListStr := ""
	for i, idea := range ideas {
		ideasListStr += fmt.Sprintf("%d. \"%s\"\n", i+1, idea)
	}

	// Build purpose text
	purposeText := purposeType
	if purpose != "" {
		purposeText += ` (` + purpose + `)`
	}

	engineeringPrompt := `
Você é um especialista em engenharia de prompts com conhecimento profundo em técnicas de prompt engineering. Sua tarefa é transformar ideias brutas e desorganizadas em um prompt estruturado, profissional e eficaz.

CONTEXTO: O usuário inseriu as seguintes notas/ideias brutas:
` + ideasListStr + `

PROPÓSITO DO PROMPT: ` + purposeText + `
IDIOMA DE RESPOSTA: ` + lang + `
TAMANHO MÁXIMO: ` + fmt.Sprintf("%d", maxLength) + ` caracteres

INSTRUÇÕES PARA ESTRUTURAÇÃO:
1. **Análise**: Identifique o objetivo principal e temas centrais das ideias
2. **Organização**: Estruture logicamente as informações em hierarquia clara
3. **Técnicas de Prompt Engineering**:
   - Definição clara de contexto e papel (persona)
   - Instruções específicas, mensuráveis e testáveis
   - Exemplos concretos quando apropriado
   - Formato de saída bem definido e estruturado
   - Chain-of-thought para raciocínios complexos
   - Few-shot examples se necessário
4. **Formatação**: Use markdown para estruturação visual clara
5. **Tom**: Seja preciso, objetivo, profissional e acionável
6. **Escopo**: Mantenha dentro do limite de caracteres especificado

CRITÉRIOS DE QUALIDADE:
- Clareza: Instruções inequívocas e fáceis de seguir
- Completude: Cubra todos os aspectos relevantes das ideias originais
- Eficácia: Optimize para obter os melhores resultados da IA
- Usabilidade: Pronto para uso imediato sem ajustes

IMPORTANTE: Responda APENAS com o prompt estruturado em markdown, sem explicações adicionais, metadados ou texto introdutório. O prompt deve ser completo, autossuficiente e pronto para uso.
`

	return engineeringPrompt
}
