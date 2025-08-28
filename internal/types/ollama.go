package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type OllamaAPI struct{ *APIConfig }

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func NewOllamaAPI(baseURL string) IAPIConfig {
	return &OllamaAPI{
		APIConfig: &APIConfig{
			baseURL: baseURL,
			httpClient: &http.Client{
				Timeout: 60 * time.Second,
			},
		},
	}
}

func (o *OllamaAPI) Complete(prompt string, stream int, model string) (string, error) {
	endpoint := fmt.Sprintf("%s/api/generate", o.baseURL)

	requestBody := OllamaRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("erro ao serializar request: %v", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("erro ao criar request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro na requisição para Ollama: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erro ao ler resposta: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama retornou status %d: %s", resp.StatusCode, string(body))
	}

	var response OllamaResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("erro ao decodificar resposta: %v", err)
	}

	return response.Response, nil
}

func (o *OllamaAPI) IsAvailable() bool {
	endpoint := fmt.Sprintf("%s/api/tags", o.baseURL)

	resp, err := o.httpClient.Get(endpoint)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func (o *OllamaAPI) GetCommonModels() []string {
	return []string{
		"ollama-chat",
		"ollama-coder",
		"ollama-math",
		"ollama-reasoner",
	}
}

func (o *OllamaAPI) ListModels() ([]string, error) {
	return []string{
		"ollama-chat",
		"ollama-coder",
		"ollama-math",
		"ollama-reasoner",
	}, nil
}

func (o *OllamaAPI) GetVersion() string { return o.version }

func (o *OllamaAPI) IsDemoMode() bool { return o.demoMode }

func (o *OllamaAPI) StartStream(string, int, string) (string, error) {
	// Implementar lógica de streaming aqui
	return "", nil
}

func (o *OllamaAPI) StopStream() error {
	// Implementar lógica para parar o streaming aqui
	return nil
}
