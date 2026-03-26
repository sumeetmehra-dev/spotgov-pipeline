package dados

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestClient_Fetch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		resp := DatasetListResponse{
			Data: []Dataset{
				{
					ID:          "abc123",
					Title:       "Contratos Públicos 2024",
					Description: "Public procurement contracts dataset",
					CreatedAt:   "2024-03-01T10:00:00",
					Organization: &Organization{
						ID:   "org1",
						Name: "Câmara Municipal de Porto",
					},
				},
			},
			Total: 1,
			Page:  1,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	client := NewClient(server.URL, logger)

	tenders, err := client.Fetch(context.Background(), time.Now().AddDate(0, -1, 0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tenders) != 1 {
		t.Fatalf("expected 1 tender, got %d", len(tenders))
	}

	tender := tenders[0]
	if tender.Source != "dados" {
		t.Errorf("expected source 'dados', got %q", tender.Source)
	}
	if tender.SourceID != "dados-abc123" {
		t.Errorf("expected source_id 'dados-abc123', got %q", tender.SourceID)
	}
	if tender.BuyerName != "Câmara Municipal de Porto" {
		t.Errorf("expected buyer 'Câmara Municipal de Porto', got %q", tender.BuyerName)
	}
	if tender.BuyerCountry != "PT" {
		t.Errorf("expected country 'PT', got %q", tender.BuyerCountry)
	}
}

func TestNormalize(t *testing.T) {
	dataset := Dataset{
		ID:          "test-001",
		Title:       "Test Dataset",
		Description: "A test procurement dataset",
		CreatedAt:   "2024-06-15",
		Organization: &Organization{
			Name: "Test Org",
		},
	}

	tender := Normalize(dataset)

	if tender.SourceID != "dados-test-001" {
		t.Errorf("expected 'dados-test-001', got %q", tender.SourceID)
	}
	if tender.Title != "Test Dataset" {
		t.Errorf("expected title 'Test Dataset', got %q", tender.Title)
	}
	if tender.PublicationDate == nil {
		t.Error("expected publication_date to be parsed")
	}
}
