package formatter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Formatter struct {
	apiKey string
	model  string
}

type chatRequest struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	Temperature float64   `json:"temperature"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func NewFormatter(apiKey, model string) *Formatter {
	return &Formatter{
		apiKey: apiKey,
		model:  model,
	}
}

// FormatResult takes the raw data and user prompt, and returns a natural language summary
func (f *Formatter) FormatResult(ctx context.Context, userPrompt string, rawData []map[string]interface{}) (string, error) {
	dataJSON, err := json.Marshal(rawData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal data: %w", err)
	}

	systemPrompt := `You are a helpful data assistant. 
Your task is to take the raw JSON data from a database query and the user's original question, 
and provide a concise, natural language answer. 
If the data is empty, state that no results were found.`

	userContent := fmt.Sprintf("Question: %s\n\nData: %s", userPrompt, string(dataJSON))

	return f.callLLM(ctx, systemPrompt, userContent)
}

func (f *Formatter) callLLM(ctx context.Context, system, user string) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"

	payload := chatRequest{
		Model: f.model,
		Messages: []message{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature: 0.7, // Slightly higher temperature for natural language generation
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+f.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("LLM API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", err
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned from LLM")
	}

	return chatResp.Choices[0].Message.Content, nil
}
