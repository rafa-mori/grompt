package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type DeepSeekAPI struct{ *APIConfig }

type DeepSeekRequest struct {
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens"`
	Model     string `json:"model"`
}

type DeepSeekAPIRequest struct {
	Model       string            `json:"model"`
	Messages    []DeepSeekMessage `json:"messages"`
	MaxTokens   int               `json:"max_tokens"`
	Temperature float64           `json:"temperature"`
	TopP        float64           `json:"top_p"`
	Stream      bool              `json:"stream"`
}

type DeepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type DeepSeekAPIResponse struct {
	ID      string           `json:"id"`
	Object  string           `json:"object"`
	Created int64            `json:"created"`
	Model   string           `json:"model"`
	Choices []DeepSeekChoice `json:"choices"`
	Usage   DeepSeekUsage    `json:"usage"`
}

type DeepSeekChoice struct {
	Index        int             `json:"index"`
	Message      DeepSeekMessage `json:"message"`
	FinishReason string          `json:"finish_reason"`
}

type DeepSeekUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type DeepSeekErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
		Param   string `json:"param"`
	} `json:"error"`
}

func NewDeepSeekAPI(apiKey string) IAPIConfig {
	return &DeepSeekAPI{
		APIConfig: &APIConfig{
			apiKey:  apiKey,
			baseURL: "https://api.deepseek.com/chat/completions",
			httpClient: &http.Client{
				Timeout: 60 * time.Second,
			},
		},
	}
}

func (d *DeepSeekAPI) Complete(prompt string, maxTokens int, model string) (string, error) {
	if d.apiKey == "" {
		return "", fmt.Errorf("API key não configurada")
	}

	// Definir modelo padrão se não especificado
	if model == "" {
		model = "deepseek-chat"
	}

	requestBody := DeepSeekAPIRequest{
		Model: model,
		Messages: []DeepSeekMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   maxTokens,
		Temperature: 0.7,
		TopP:        0.95,
		Stream:      false,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("erro ao serializar request: %v", err)
	}

	req, err := http.NewRequest("POST", d.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("erro ao criar request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.apiKey)
	req.Header.Set("User-Agent", "PromptCrafter/1.0")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro na requisição: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao ler resposta: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Error handling for non-200 responses
		var errorResp DeepSeekErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return "", fmt.Errorf("DeepSeek API erro: %s", errorResp.Error.Message)
		}
		return "", fmt.Errorf("API retornou status %d: %s", resp.StatusCode, string(body))
	}

	var response DeepSeekAPIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("resposta vazia da API")
	}

	return response.Choices[0].Message.Content, nil
}

func (d *DeepSeekAPI) IsAvailable() bool {
	if d.apiKey == "" {
		return false
	}

	// Make a simple request to check API availability
	testReq := DeepSeekAPIRequest{
		Model: "deepseek-chat",
		Messages: []DeepSeekMessage{
			{
				Role:    "user",
				Content: "test",
			},
		},
		MaxTokens: 1,
	}

	jsonData, err := json.Marshal(testReq)
	if err != nil {
		return false
	}

	req, err := http.NewRequest("POST", d.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return false
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// If the status code is 200 OK or 400 Bad Request, we consider the API available
	return resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest
}

// GetAvailableModels Available DeepSeek models
func (d *DeepSeekAPI) GetAvailableModels() []string {
	return []string{
		"deepseek-chat",
		"deepseek-coder",
		"deepseek-math",
		"deepseek-reasoner",
	}
}

// HealthCheck Lightweight health check to verify API key and basic connectivity
func (d *DeepSeekAPI) HealthCheck() error {
	if d.apiKey == "" {
		return fmt.Errorf("API key não configurada")
	}

	// Make a simple request to check API key validity
	req, err := http.NewRequest("GET", "https://api.deepseek.com/models", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+d.apiKey)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("API key inválida")
	}

	if resp.StatusCode >= 500 {
		return fmt.Errorf("servidor DeepSeek indisponível")
	}

	return nil
}

// EstimateCost Estimates the cost of a request based on prompt and completion tokens
func (d *DeepSeekAPI) EstimateCost(promptTokens, completionTokens int, model string) float64 {
	var promptPrice, completionPrice float64

	switch model {
	case "deepseek-chat":
		promptPrice = 0.14 / 1000000     // $0.14 by 1M prompt tokens
		completionPrice = 0.28 / 1000000 // $0.28 by 1M completion tokens
	case "deepseek-coder":
		promptPrice = 0.14 / 1000000
		completionPrice = 0.28 / 1000000
	case "deepseek-math":
		promptPrice = 0.14 / 1000000
		completionPrice = 0.28 / 1000000
	default:
		promptPrice = 0.14 / 1000000
		completionPrice = 0.28 / 1000000
	}

	return float64(promptTokens)*promptPrice + float64(completionTokens)*completionPrice
}

// GetVersion returns the version of the API
func (d *DeepSeekAPI) GetVersion() string { return d.version }

// IsDemoMode indicates if the API is in demo mode
func (d *DeepSeekAPI) IsDemoMode() bool { return d.demoMode }

// ListModels retrieves the list of available models from the DeepSeek API
func (d *DeepSeekAPI) ListModels() ([]string, error) {
	return []string{
		"deepseek-chat",
		"deepseek-coder",
		"deepseek-math",
		"deepseek-reasoner",
	}, nil
}

// GetCommonModels retrieves the list of common models from the DeepSeek API
func (d *DeepSeekAPI) GetCommonModels() []string {
	return []string{
		"deepseek-chat",
		"deepseek-coder",
		"deepseek-math",
		"deepseek-reasoner",
	}
}

func (d *DeepSeekAPI) StartStream(string, int, string) (string, error) {
	// Implementar lógica de streaming aqui
	return "", nil
}

func (d *DeepSeekAPI) StopStream() error {
	// Implementar lógica para parar o streaming aqui
	return nil
}
