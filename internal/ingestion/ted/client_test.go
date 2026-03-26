package ted

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestClient_Search(t *testing.T) {
	tests := []struct {
		name       string
		response   SearchResponse
		statusCode int
		wantCount  int
		wantErr    bool
	}{
		{
			name: "successful search with results",
			response: SearchResponse{
				Notices: []Notice{
					{
						PublicationNumber: "2024/S 001-000001",
						Title:            "Construction of highway bridge",
						BuyerCountry:     "PT",
						CPVCodes:         []string{"45000000"},
					},
					{
						PublicationNumber: "2024/S 001-000002",
						Title:            "IT consulting services",
						BuyerCountry:     "PT",
						CPVCodes:         []string{"72000000"},
					},
				},
				Total: 2,
				Pages: 1,
				Page:  1,
			},
			statusCode: http.StatusOK,
			wantCount:  2,
			wantErr:    false,
		},
		{
			name:       "empty results",
			response:   SearchResponse{Notices: []Notice{}, Total: 0, Pages: 0, Page: 1},
			statusCode: http.StatusOK,
			wantCount:  0,
			wantErr:    false,
		},
		{
			name:       "server error",
			response:   SearchResponse{},
			statusCode: http.StatusInternalServerError,
			wantCount:  0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != searchPath {
					t.Errorf("expected path %s, got %s", searchPath, r.URL.Path)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			logger, _ := zap.NewDevelopment()
			client := NewClient(server.URL, logger)

			tenders, err := client.Fetch(context.Background(), time.Now().AddDate(0, -1, 0))
			if (err != nil) != tt.wantErr {
				t.Errorf("Fetch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(tenders) != tt.wantCount {
				t.Errorf("Fetch() returned %d tenders, want %d", len(tenders), tt.wantCount)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	value := 1500000.0
	notice := Notice{
		PublicationNumber:  "2024/S 050-123456",
		Title:              "Construction of municipal road",
		Description:        "Road construction in Lisbon municipality",
		CPVCodes:           []string{"45233120", "45233140"},
		BuyerName:          "Câmara Municipal de Lisboa",
		BuyerCountry:       "PT",
		PlaceOfPerformance: "Lisbon, Portugal",
		ProcurementMethod:  "open",
		EstimatedValue:     &value,
		Currency:           "EUR",
		Deadline:           "2024-06-15",
		PublicationDate:    "2024-03-01",
		NoticeURL:          "https://ted.europa.eu/udl?uri=TED:NOTICE:123456-2024:TEXT:EN:HTML",
	}

	tender := Normalize(notice)

	if tender.Source != "ted" {
		t.Errorf("expected source 'ted', got %q", tender.Source)
	}
	if tender.SourceID != notice.PublicationNumber {
		t.Errorf("expected source_id %q, got %q", notice.PublicationNumber, tender.SourceID)
	}
	if tender.Title != notice.Title {
		t.Errorf("expected title %q, got %q", notice.Title, tender.Title)
	}
	if len(tender.CPVCodes) != 2 {
		t.Errorf("expected 2 CPV codes, got %d", len(tender.CPVCodes))
	}
	if tender.Deadline == nil {
		t.Error("expected deadline to be parsed")
	}
	if tender.PublicationDate == nil {
		t.Error("expected publication_date to be parsed")
	}
	if *tender.EstimatedValue != value {
		t.Errorf("expected value %f, got %f", value, *tender.EstimatedValue)
	}
	if tender.Currency != "EUR" {
		t.Errorf("expected currency EUR, got %q", tender.Currency)
	}
}

func TestNormalize_MissingFields(t *testing.T) {
	notice := Notice{
		PublicationNumber: "2024/S 001-000001",
		Title:             "Minimal notice",
	}

	tender := Normalize(notice)

	if tender.SourceID != notice.PublicationNumber {
		t.Errorf("expected source_id %q, got %q", notice.PublicationNumber, tender.SourceID)
	}
	if tender.Currency != "EUR" {
		t.Errorf("expected default currency EUR, got %q", tender.Currency)
	}
	if tender.Deadline != nil {
		t.Error("expected nil deadline for empty date")
	}
	if tender.EstimatedValue != nil {
		t.Error("expected nil estimated value")
	}
}
