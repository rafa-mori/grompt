package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ClaudeAPI struct{ *APIConfig }

type ClaudeRequest struct {
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens"`
}

type ClaudeAPIRequest struct {
	Model     string `json:"model"`
	MaxTokens int    `json:"max_tokens"`
	Messages  []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

type ClaudeAPIResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

func NewClaudeAPI(apiKey string) IAPIConfig {
	return &ClaudeAPI{
		APIConfig: &APIConfig{
			apiKey:  apiKey,
			baseURL: "https://api.anthropic.com/v1/messages",
			httpClient: &http.Client{
				Timeout: 30 * time.Second,
			},
		},
	}
}

// Complete sends a completion request to the Claude API
func (c *ClaudeAPI) Complete(prompt string, maxTokens int, model string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("API key não configurada")
	}

	requestBody := ClaudeAPIRequest{
		Model:     "claude-3-sonnet-20240229",
		MaxTokens: maxTokens,
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("erro ao serializar request: %v", err)
	}

	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("erro ao criar request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro na requisição: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao ler resposta: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API retornou status %d: %s", resp.StatusCode, string(body))
	}

	var response ClaudeAPIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	if len(response.Content) == 0 {
		return "", fmt.Errorf("resposta vazia da API")
	}

	return response.Content[0].Text, nil
}

// IsAvailable checks if the Claude API is available
func (c *ClaudeAPI) IsAvailable() bool {
	if c.apiKey == "" {
		return false
	}

	// Make a simple request to check API availability
	testReq := ClaudeAPIRequest{
		Model: "claude-3-sonnet-20240229",
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
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

	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return false
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// If the status code is 200 OK or 400 Bad Request, we consider the API available
	return resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest
}

// ListModels retrieves the list of available models from the Claude API
func (c *ClaudeAPI) ListModels() ([]string, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("API key não configurada")
	}

	req, err := http.NewRequest("GET", c.baseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro na requisição: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API retornou status %d: %s", resp.StatusCode, string(body))
	}

	var models []string
	if err := json.Unmarshal(body, &models); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	return models, nil
}

// GetCommonModels retrieves the list of common models from the Claude API
func (c *ClaudeAPI) GetCommonModels() []string {
	return []string{
		"claude-3-sonnet-20240229",
		"claude-3-5-sonnet-20240229",
	}
}

// GetVersion returns the version of the API
func (c *ClaudeAPI) GetVersion() string { return c.version }

// IsDemoMode indicates if the API is in demo mode
func (c *ClaudeAPI) IsDemoMode() bool { return c.demoMode }

func (c *ClaudeAPI) StartStream(string, int, string) (string, error) {
	// Implementar lógica de streaming aqui
	return "", nil
}

func (c *ClaudeAPI) StopStream() error {
	// Implementar lógica para parar o streaming aqui
	return nil
}
