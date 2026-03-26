package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/sumeetmehra/procura/internal/model"
	"go.uber.org/zap"
)

// Indexer handles indexing tenders into Elasticsearch.
type Indexer struct {
	es     *ESClient
	logger *zap.Logger
}

func NewIndexer(es *ESClient, logger *zap.Logger) *Indexer {
	return &Indexer{
		es:     es,
		logger: logger.Named("indexer"),
	}
}

// IndexTender indexes a single tender document.
func (idx *Indexer) IndexTender(ctx context.Context, tender model.Tender) error {
	doc := map[string]interface{}{
		"id":                   tender.ID.String(),
		"source":               tender.Source,
		"source_id":            tender.SourceID,
		"title":                tender.Title,
		"description":          tender.Description,
		"cpv_codes":            tender.CPVCodes,
		"buyer_name":           tender.BuyerName,
		"buyer_country":        tender.BuyerCountry,
		"place_of_performance": tender.PlaceOfPerformance,
		"procurement_method":   tender.ProcurementMethod,
		"estimated_value":      tender.EstimatedValue,
		"currency":             tender.Currency,
		"deadline":             tender.Deadline,
		"publication_date":     tender.PublicationDate,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(doc); err != nil {
		return fmt.Errorf("encode document: %w", err)
	}

	res, err := idx.es.client.Index(
		indexName,
		&buf,
		idx.es.client.Index.WithContext(ctx),
		idx.es.client.Index.WithDocumentID(tender.ID.String()),
	)
	if err != nil {
		return fmt.Errorf("index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("index error: %s", string(body))
	}

	return nil
}

// BulkIndex indexes multiple tenders using the bulk API.
func (idx *Indexer) BulkIndex(ctx context.Context, tenders []model.Tender) (int, error) {
	if len(tenders) == 0 {
		return 0, nil
	}

	var buf bytes.Buffer
	indexed := 0

	for _, tender := range tenders {
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": indexName,
				"_id":    tender.ID.String(),
			},
		}
		if err := json.NewEncoder(&buf).Encode(meta); err != nil {
			return indexed, fmt.Errorf("encode meta: %w", err)
		}

		doc := map[string]interface{}{
			"id":                   tender.ID.String(),
			"source":               tender.Source,
			"source_id":            tender.SourceID,
			"title":                tender.Title,
			"description":          tender.Description,
			"cpv_codes":            tender.CPVCodes,
			"buyer_name":           tender.BuyerName,
			"buyer_country":        tender.BuyerCountry,
			"place_of_performance": tender.PlaceOfPerformance,
			"procurement_method":   tender.ProcurementMethod,
			"estimated_value":      tender.EstimatedValue,
			"currency":             tender.Currency,
			"deadline":             tender.Deadline,
			"publication_date":     tender.PublicationDate,
		}
		if err := json.NewEncoder(&buf).Encode(doc); err != nil {
			return indexed, fmt.Errorf("encode doc: %w", err)
		}

		indexed++
	}

	res, err := idx.es.client.Bulk(
		bytes.NewReader(buf.Bytes()),
		idx.es.client.Bulk.WithContext(ctx),
		idx.es.client.Bulk.WithIndex(indexName),
	)
	if err != nil {
		return 0, fmt.Errorf("bulk index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return 0, fmt.Errorf("bulk index error: %s", string(body))
	}

	idx.logger.Info("bulk indexed tenders", zap.Int("count", indexed))
	return indexed, nil
}
