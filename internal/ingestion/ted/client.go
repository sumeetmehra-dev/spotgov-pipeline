package ted

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sumeetmehra/spotgov-pipeline/internal/ingestion"
	"github.com/sumeetmehra/spotgov-pipeline/internal/model"
	"go.uber.org/zap"
)

const (
	searchPath     = "/v3/notices/search"
	maxPerPage     = 100
	requestTimeout = 30 * time.Second
)

// responseFields are the TED v3 fields we request in every search.
var responseFields = []string{
	"publication-number",
	"notice-type",
	"publication-date",
	"title-proc",
	"description-lot",
	"description-proc",
	"organisation-name-buyer",
	"organisation-country-buyer",
	"estimated-value-lot",
	"estimated-value-cur-lot",
	"deadline-receipt-tender-date-lot",
	"classification-cpv",
	"contract-nature",
}

// Client communicates with the TED Search API v3.
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// Ensure Client implements Source.
var _ ingestion.Source = (*Client)(nil)

func NewClient(baseURL string, logger *zap.Logger) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
		logger: logger.Named("ted"),
	}
}

func (c *Client) Name() string {
	return "ted"
}

// Fetch retrieves Portuguese procurement notices published after the given time.
func (c *Client) Fetch(ctx context.Context, since time.Time) ([]model.Tender, error) {
	var allTenders []model.Tender
	page := 1

	sinceStr := since.Format("20060102")
	query := fmt.Sprintf(
		"buyer-country = PRT AND notice-type = cn-standard AND publication-date >= %s SORT BY publication-date DESC",
		sinceStr,
	)

	for {
		req := SearchRequest{
			Query:          query,
			Fields:         responseFields,
			Limit:          maxPerPage,
			Page:           page,
			PaginationMode: "PAGE_NUMBER",
			Scope:          "ACTIVE",
		}

		resp, err := c.search(ctx, req)
		if err != nil {
			return allTenders, fmt.Errorf("search page %d: %w", page, err)
		}

		c.logger.Info("fetched TED notices",
			zap.Int("page", page),
			zap.Int("count", len(resp.Notices)),
			zap.Int("total", resp.TotalNoticeCount),
		)

		for _, notice := range resp.Notices {
			tender := Normalize(notice)
			allTenders = append(allTenders, tender)
		}

		// Stop when we've fetched all pages or got empty results.
		totalPages := (resp.TotalNoticeCount + maxPerPage - 1) / maxPerPage
		if page >= totalPages || len(resp.Notices) == 0 {
			break
		}
		page++

		// Rate limiting: avoid hammering the API.
		select {
		case <-ctx.Done():
			return allTenders, ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
	}

	c.logger.Info("TED ingestion complete", zap.Int("total_tenders", len(allTenders)))
	return allTenders, nil
}

func (c *Client) search(ctx context.Context, searchReq SearchRequest) (*SearchResponse, error) {
	body, err := json.Marshal(searchReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := c.baseURL + searchPath
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TED API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &searchResp, nil
}
