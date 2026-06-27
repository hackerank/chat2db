package generator

import (
	"bytes"
	"chat2db/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Generator struct {
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

func NewGenerator(apiKey, model string) *Generator {
	return &Generator{
		apiKey: apiKey,
		model:  model,
	}
}

// GenerateSQL takes the schema and user prompt, and returns a raw SQL string
func (g *Generator) GenerateSQL(ctx context.Context, schema models.DatabaseSchema, userPrompt string) (string, error) {
	// 1. Serialize schema to JSON for the prompt
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to serialize schema: %w", err)
	}

	// 2. Construct the prompt
	systemPrompt := fmt.Sprintf(`You are a MySQL expert. 
Your task is to generate a valid, read-only MySQL query based on the provided database schema.
Return ONLY the raw SQL query. Do not include markdown formatting, code blocks, or explanations.

Database Schema:
%s`, string(schemaJSON))

	// 3. Call the LLM
	sql, err := g.callLLM(ctx, systemPrompt, userPrompt)
	if err != nil {
		return "", err
	}

	// 4. Clean the output (remove potential markdown artifacts)
	cleaned := strings.TrimSpace(sql)
	cleaned = strings.ReplaceAll(cleaned, "```sql", "")
	cleaned = strings.ReplaceAll(cleaned, "```", "")
	
	return strings.TrimSpace(cleaned), nil
}

func (g *Generator) callLLM(ctx context.Context, system, user string) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"

	payload := chatRequest{
		Model: g.model,
		Messages: []message{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature: 0,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+g.apiKey)
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
