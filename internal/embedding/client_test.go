package embedding

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestClient_EmbedTexts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Error("missing or wrong authorization header")
		}

		var req embeddingRequest
		json.NewDecoder(r.Body).Decode(&req)

		resp := embeddingResponse{
			Data: make([]embeddingData, len(req.Input)),
		}
		for i := range req.Input {
			vec := make(Vector, 4)
			for j := range vec {
				vec[j] = float64(i+1) * 0.1
			}
			resp.Data[i] = embeddingData{
				Embedding: vec,
				Index:     i,
			}
		}
		resp.Usage.TotalTokens = len(req.Input) * 10

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	client := &Client{
		apiKey:     "test-key",
		model:      "mistral-embed",
		dimensions: 4,
		httpClient: &http.Client{},
		url:        server.URL,
		logger:     logger,
	}

	vectors, err := client.EmbedTexts(context.Background(), []string{"hello world", "test text"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vectors) != 2 {
		t.Fatalf("expected 2 vectors, got %d", len(vectors))
	}
	if len(vectors[0]) != 4 {
		t.Errorf("expected vector dim 4, got %d", len(vectors[0]))
	}
}

func TestClient_EmbedText_Single(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := embeddingResponse{
			Data: []embeddingData{
				{Embedding: Vector{0.1, 0.2, 0.3}, Index: 0},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	client := &Client{
		apiKey:     "test-key",
		model:      "mistral-embed",
		dimensions: 3,
		httpClient: &http.Client{},
		url:        server.URL,
		logger:     logger,
	}

	vec, err := client.EmbedText(context.Background(), "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vec) != 3 {
		t.Errorf("expected dim 3, got %d", len(vec))
	}
}

func TestClient_EmptyInput(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	client := &Client{
		apiKey: "test-key",
		logger: logger,
	}

	vectors, err := client.EmbedTexts(context.Background(), []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vectors != nil {
		t.Errorf("expected nil for empty input, got %v", vectors)
	}
}

func TestNewClient_MissingKey(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_, err := NewClient("", logger)
	if err == nil {
		t.Error("expected error for missing API key")
	}
}
