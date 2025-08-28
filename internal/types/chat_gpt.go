package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ChatGPTAPI struct{ *APIConfig }

type ChatGPTRequest struct {
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens"`
	Model     string `json:"model"`
}

type ChatGPTAPIRequest struct {
	Model       string           `json:"model"`
	Messages    []ChatGPTMessage `json:"messages"`
	MaxTokens   int              `json:"max_tokens"`
	Temperature float64          `json:"temperature"`
	Stream      bool             `json:"stream"`
}

type ChatGPTMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatGPTAPIResponse struct {
	ID      string          `json:"id"`
	Object  string          `json:"object"`
	Created int64           `json:"created"`
	Model   string          `json:"model"`
	Choices []ChatGPTChoice `json:"choices"`
	Usage   ChatGPTUsage    `json:"usage"`
}

type ChatGPTChoice struct {
	Index        int            `json:"index"`
	Message      ChatGPTMessage `json:"message"`
	FinishReason string         `json:"finish_reason"`
}

type ChatGPTUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatGPTErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

func NewChatGPTAPI(apiKey string) IAPIConfig {
	return &ChatGPTAPI{
		APIConfig: &APIConfig{
			apiKey:  apiKey,
			baseURL: "https://api.chatgpt.com/v1/chat/completions",
			httpClient: &http.Client{
				Timeout: 60 * time.Second,
			},
		},
	}
}

func (o *ChatGPTAPI) Complete(prompt string, maxTokens int, model string) (string, error) {
	if o.apiKey == "" {
		return "", fmt.Errorf("API key não configurada")
	}

	// Definir modelo padrão se não especificado
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	requestBody := ChatGPTAPIRequest{
		Model: model,
		Messages: []ChatGPTMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens:   maxTokens,
		Temperature: 0.7,
		Stream:      false,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("erro ao serializar request: %v", err)
	}

	req, err := http.NewRequest("POST", o.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("erro ao criar request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro na requisição: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao ler resposta: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Tentar parsear erro da ChatGPT
		var errorResp ChatGPTErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return "", fmt.Errorf("ChatGPT API erro: %s", errorResp.Error.Message)
		}
		return "", fmt.Errorf("API retornou status %d: %s", resp.StatusCode, string(body))
	}

	var response ChatGPTAPIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("resposta vazia da API")
	}

	return response.Choices[0].Message.Content, nil
}

func (o *ChatGPTAPI) IsAvailable() bool {
	if o.apiKey == "" {
		return false
	}

	// Fazer uma requisição simples para verificar se a API está funcionando
	testReq := ChatGPTAPIRequest{
		Model: "gpt-3.5-turbo",
		Messages: []ChatGPTMessage{
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

	req, err := http.NewRequest("POST", o.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return false
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Se retornou 200 ou 400 (bad request), significa que a API está respondendo
	// 401 significa unauthorized (API key inválida)
	// 429 significa rate limit
	return resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusBadRequest
}

// Listar modelos disponíveis
func (o *ChatGPTAPI) ListModels() ([]string, error) {
	if o.apiKey == "" {
		return nil, fmt.Errorf("API key não configurada")
	}

	modelsURL := "https://api.chatgpt.com/v1/models"
	req, err := http.NewRequest("GET", modelsURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("erro ao listar modelos: status %d", resp.StatusCode)
	}

	type ChatGPTModelsResponse struct {
		Data []struct {
			ID     string `json:"id"`
			Object string `json:"object"`
		} `json:"data"`
	}

	var modelsResp ChatGPTModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, err
	}

	var models []string
	for _, model := range modelsResp.Data {
		// Filtrar apenas modelos de chat
		if model.Object == "model" {
			models = append(models, model.ID)
		}
	}

	return models, nil
}

// Modelos comuns da ChatGPT
func (o *ChatGPTAPI) GetCommonModels() []string {
	return []string{
		"gpt-4",
		"gpt-4-turbo",
		"gpt-4o",
		"gpt-4o-mini",
		"gpt-3.5-turbo",
		"gpt-3.5-turbo-16k",
	}
}

// GetVersion returns the version of the API
func (o *ChatGPTAPI) GetVersion() string { return o.version }

// IsDemoMode indicates if the API is in demo mode
func (o *ChatGPTAPI) IsDemoMode() bool { return o.demoMode }

func (o *ChatGPTAPI) StartStream(string, int, string) (string, error) {
	// Implementar lógica de streaming aqui
	return "", nil
}

func (o *ChatGPTAPI) StopStream() error {
	// Implementar lógica para parar o streaming aqui
	return nil
}
