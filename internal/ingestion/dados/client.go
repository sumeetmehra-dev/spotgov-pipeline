package dados

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sumeetmehra/spotgov-pipeline/internal/ingestion"
	"github.com/sumeetmehra/spotgov-pipeline/internal/model"
	"go.uber.org/zap"
)

const requestTimeout = 30 * time.Second

// Client communicates with the dados.gov.pt API to fetch Portuguese procurement data.
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

var _ ingestion.Source = (*Client)(nil)

func NewClient(baseURL string, logger *zap.Logger) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
		logger: logger.Named("dados"),
	}
}

func (c *Client) Name() string {
	return "dados"
}

// Fetch retrieves procurement datasets from dados.gov.pt and normalizes them.
func (c *Client) Fetch(ctx context.Context, since time.Time) ([]model.Tender, error) {
	url := fmt.Sprintf("%s/datasets/?q=contratos+publicos&page_size=50", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dados API returned status %d", resp.StatusCode)
	}

	var listResp DatasetListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	c.logger.Info("fetched dados.gov.pt datasets",
		zap.Int("count", len(listResp.Data)),
		zap.Int("total", listResp.Total),
	)

	var tenders []model.Tender
	for _, dataset := range listResp.Data {
		tender := Normalize(dataset)
		if tender.SourceID != "" {
			tenders = append(tenders, tender)
		}
	}

	c.logger.Info("dados.gov.pt ingestion complete", zap.Int("tenders", len(tenders)))
	return tenders, nil
}
