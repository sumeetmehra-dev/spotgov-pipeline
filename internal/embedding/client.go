package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

const (
	defaultMistralURL = "https://api.mistral.ai/v1/embeddings"
	defaultModel      = "mistral-embed"
	defaultDimensions = 1024
)

// Vector is a float64 slice representing an embedding vector.
type Vector []float64

// Client generates embeddings via the Mistral AI API.
type Client struct {
	apiKey     string
	model      string
	dimensions int
	httpClient *http.Client
	url        string
	logger     *zap.Logger
}

func NewClient(apiKey string, logger *zap.Logger) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("Mistral API key required: set MISTRAL_API_KEY or pass via config")
	}

	return &Client{
		apiKey:     apiKey,
		model:      defaultModel,
		dimensions: defaultDimensions,
		httpClient: &http.Client{},
		url:        defaultMistralURL,
		logger:     logger.Named("embedding"),
	}, nil
}

// EmbedTexts generates embeddings for a batch of texts.
func (c *Client) EmbedTexts(ctx context.Context, texts []string) ([]Vector, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	reqBody := embeddingRequest{
		Input:      texts,
		Model:      c.model,
		Dimensions: c.dimensions,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call Mistral API: %w", err)
	}
	defer resp.Body.Close()

	var result embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("Mistral API error: %s", result.Error.Message)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Mistral API returned status %d", resp.StatusCode)
	}

	vectors := make([]Vector, len(result.Data))
	for _, d := range result.Data {
		vectors[d.Index] = d.Embedding
	}

	c.logger.Debug("generated embeddings",
		zap.Int("count", len(vectors)),
		zap.Int("tokens", result.Usage.TotalTokens),
	)

	return vectors, nil
}

// EmbedText generates an embedding for a single text.
func (c *Client) EmbedText(ctx context.Context, text string) (Vector, error) {
	vectors, err := c.EmbedTexts(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(vectors) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return vectors[0], nil
}

type embeddingRequest struct {
	Input      []string `json:"input"`
	Model      string   `json:"model"`
	Dimensions int      `json:"dimensions,omitempty"`
}

type embeddingResponse struct {
	Data  []embeddingData `json:"data"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type embeddingData struct {
	Embedding Vector `json:"embedding"`
	Index     int    `json:"index"`
}
