package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type OpenAIAPI struct{ *APIConfig }

type OpenAIRequest struct {
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens"`
	Model     string `json:"model"`
}

type OpenAIAPIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
	Stream      bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIAPIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type OpenAIErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

func NewOpenAIAPI(apiKey string) IAPIConfig {
	return &OpenAIAPI{
		APIConfig: &APIConfig{
			apiKey:  apiKey,
			baseURL: "https://api.openai.com/v1/chat/completions",
			httpClient: &http.Client{
				Timeout: 60 * time.Second,
			},
		},
	}
}

func (o *OpenAIAPI) Complete(prompt string, maxTokens int, model string) (string, error) {
	if o.apiKey == "" {
		return "", fmt.Errorf("API key não configurada")
	}

	// Definir modelo padrão se não especificado
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	requestBody := OpenAIAPIRequest{
		Model: model,
		Messages: []Message{
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
		// Tentar parsear erro da OpenAI
		var errorResp OpenAIErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return "", fmt.Errorf("OpenAI API erro: %s", errorResp.Error.Message)
		}
		return "", fmt.Errorf("API retornou status %d: %s", resp.StatusCode, string(body))
	}

	var response OpenAIAPIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("resposta vazia da API")
	}

	return response.Choices[0].Message.Content, nil
}

func (o *OpenAIAPI) IsAvailable() bool {
	if o.apiKey == "" {
		return false
	}

	// Fazer uma requisição simples para verificar se a API está funcionando
	testReq := OpenAIAPIRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
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
func (o *OpenAIAPI) ListModels() ([]string, error) {
	if o.apiKey == "" {
		return nil, fmt.Errorf("API key não configurada")
	}

	modelsURL := "https://api.openai.com/v1/models"
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

	type ModelsResponse struct {
		Data []struct {
			ID     string `json:"id"`
			Object string `json:"object"`
		} `json:"data"`
	}

	var modelsResp ModelsResponse
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

// Modelos comuns da OpenAI
func (o *OpenAIAPI) GetCommonModels() []string {
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
func (o *OpenAIAPI) GetVersion() string { return o.version }

// IsDemoMode indicates if the API is in demo mode
func (o *OpenAIAPI) IsDemoMode() bool { return o.demoMode }

func (o *OpenAIAPI) StartStream(string, int, string) (string, error) {
	// Implementar lógica de streaming aqui
	return "", nil
}

func (o *OpenAIAPI) StopStream() error {
	// Implementar lógica para parar o streaming aqui
	return nil
}
