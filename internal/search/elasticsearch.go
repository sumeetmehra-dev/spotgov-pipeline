package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/sumeetmehra/procura/internal/model"
	"go.uber.org/zap"
)

const indexName = "tenders"

// ESClient wraps the Elasticsearch client with tender-specific operations.
type ESClient struct {
	client *elasticsearch.Client
	logger *zap.Logger
}

func NewESClient(url string, logger *zap.Logger) (*ESClient, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{url},
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("create ES client: %w", err)
	}

	// Verify connection
	res, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("ES connection failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ES connection error: %s", res.String())
	}

	logger.Info("connected to Elasticsearch")

	return &ESClient{
		client: client,
		logger: logger.Named("elasticsearch"),
	}, nil
}

// CreateIndex creates the tenders index with Portuguese and English analyzers.
func (e *ESClient) CreateIndex(ctx context.Context) error {
	mapping := `{
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0,
			"analysis": {
				"analyzer": {
					"pt_analyzer": {
						"type": "custom",
						"tokenizer": "standard",
						"filter": ["lowercase", "portuguese_stop", "portuguese_stemmer"]
					}
				},
				"filter": {
					"portuguese_stop": {
						"type": "stop",
						"stopwords": "_portuguese_"
					},
					"portuguese_stemmer": {
						"type": "stemmer",
						"language": "portuguese"
					}
				}
			}
		},
		"mappings": {
			"properties": {
				"id":                   {"type": "keyword"},
				"source":               {"type": "keyword"},
				"source_id":            {"type": "keyword"},
				"title": {
					"type": "text",
					"analyzer": "standard",
					"fields": {
						"pt": {"type": "text", "analyzer": "pt_analyzer"},
						"en": {"type": "text", "analyzer": "english"}
					}
				},
				"description": {
					"type": "text",
					"analyzer": "standard",
					"fields": {
						"pt": {"type": "text", "analyzer": "pt_analyzer"},
						"en": {"type": "text", "analyzer": "english"}
					}
				},
				"cpv_codes":            {"type": "keyword"},
				"buyer_name":           {"type": "text", "fields": {"keyword": {"type": "keyword"}}},
				"buyer_country":        {"type": "keyword"},
				"place_of_performance": {"type": "text", "fields": {"keyword": {"type": "keyword"}}},
				"procurement_method":   {"type": "keyword"},
				"estimated_value":      {"type": "float"},
				"currency":             {"type": "keyword"},
				"deadline":             {"type": "date"},
				"publication_date":     {"type": "date"}
			}
		}
	}`

	res, err := e.client.Indices.Exists([]string{indexName})
	if err != nil {
		return fmt.Errorf("check index existence: %w", err)
	}
	defer res.Body.Close()

	if !res.IsError() {
		e.logger.Info("index already exists", zap.String("index", indexName))
		return nil
	}

	res, err = e.client.Indices.Create(
		indexName,
		e.client.Indices.Create.WithBody(strings.NewReader(mapping)),
		e.client.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("create index error: %s", string(body))
	}

	e.logger.Info("created index", zap.String("index", indexName))
	return nil
}

// SearchFilters holds optional search filters.
type SearchFilters struct {
	CPVCode   string
	Country   string
	MinValue  *float64
	MaxValue  *float64
}

// SearchResult represents a single search hit.
type SearchResult struct {
	Tender model.Tender `json:"tender"`
	Score  float64      `json:"score"`
}

// Search performs a multi-match search across title and description with filters.
func (e *ESClient) Search(ctx context.Context, query string, filters SearchFilters, limit int) ([]SearchResult, error) {
	must := []map[string]interface{}{
		{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"title^3", "title.pt^2", "title.en^2", "description", "description.pt", "description.en", "buyer_name"},
				"type":   "best_fields",
			},
		},
	}

	filterClauses := buildFilterClauses(filters)

	body := map[string]interface{}{
		"size": limit,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must":   must,
				"filter": filterClauses,
			},
		},
		"sort": []interface{}{
			"_score",
			map[string]interface{}{"publication_date": map[string]string{"order": "desc", "missing": "_last"}},
		},
	}

	return e.executeSearch(ctx, body)
}

func (e *ESClient) executeSearch(ctx context.Context, body map[string]interface{}) ([]SearchResult, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, fmt.Errorf("encode search body: %w", err)
	}

	res, err := e.client.Search(
		e.client.Search.WithContext(ctx),
		e.client.Search.WithIndex(indexName),
		e.client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, fmt.Errorf("execute search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		respBody, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("search error: %s", string(respBody))
	}

	var result struct {
		Hits struct {
			Hits []struct {
				Score  float64                `json:"_score"`
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode search response: %w", err)
	}

	var results []SearchResult
	for _, hit := range result.Hits.Hits {
		sourceBytes, _ := json.Marshal(hit.Source)
		var tender model.Tender
		json.Unmarshal(sourceBytes, &tender)

		results = append(results, SearchResult{
			Tender: tender,
			Score:  hit.Score,
		})
	}

	return results, nil
}

func buildFilterClauses(f SearchFilters) []map[string]interface{} {
	var filters []map[string]interface{}

	if f.CPVCode != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{"cpv_codes": f.CPVCode},
		})
	}
	if f.Country != "" {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{"buyer_country": f.Country},
		})
	}
	if f.MinValue != nil || f.MaxValue != nil {
		rangeFilter := map[string]interface{}{}
		if f.MinValue != nil {
			rangeFilter["gte"] = *f.MinValue
		}
		if f.MaxValue != nil {
			rangeFilter["lte"] = *f.MaxValue
		}
		filters = append(filters, map[string]interface{}{
			"range": map[string]interface{}{"estimated_value": rangeFilter},
		})
	}

	return filters
}
