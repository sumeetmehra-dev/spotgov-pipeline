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
						PublicationNumber: "136388-2025",
						NoticeType:        "cn-standard",
						PublicationDate:   "2025-03-03+01:00",
						TitleProc:         I18nText{"por": "Construção de ponte rodoviária"},
						OrgCountryBuyer:   []string{"PRT"},
						CPVCodes:          []string{"45000000"},
					},
					{
						PublicationNumber: "136389-2025",
						NoticeType:        "cn-standard",
						PublicationDate:   "2025-03-03+01:00",
						TitleProc:         I18nText{"eng": "IT consulting services"},
						OrgCountryBuyer:   []string{"PRT"},
						CPVCodes:          []string{"72000000"},
					},
				},
				TotalNoticeCount: 2,
			},
			statusCode: http.StatusOK,
			wantCount:  2,
			wantErr:    false,
		},
		{
			name:       "empty results",
			response:   SearchResponse{Notices: []Notice{}, TotalNoticeCount: 0},
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
	notice := Notice{
		PublicationNumber: "136388-2025",
		NoticeType:        "cn-standard",
		PublicationDate:   "2025-03-03+01:00",
		TitleProc:         I18nText{"por": "Construção de estrada municipal", "eng": "Construction of municipal road"},
		DescriptionLot:    I18nTextArray{"eng": {"Road construction in Lisbon municipality"}},
		OrgNameBuyer:      I18nTextArray{"por": {"Câmara Municipal de Lisboa"}},
		OrgCountryBuyer:   []string{"PRT"},
		EstimatedValueLot: []string{"1500000"},
		EstimatedValueCur: []string{"EUR"},
		DeadlineDateLot:   []string{"2025-06-15Z"},
		CPVCodes:          []string{"45233120", "45233140"},
		ContractNature:    []string{"works"},
		Links: NoticeLinks{
			HTML: map[string]string{
				"ENG": "https://ted.europa.eu/en/notice/-/detail/136388-2025",
			},
		},
	}

	tender := Normalize(notice)

	if tender.Source != "ted" {
		t.Errorf("expected source 'ted', got %q", tender.Source)
	}
	if tender.SourceID != "136388-2025" {
		t.Errorf("expected source_id '136388-2025', got %q", tender.SourceID)
	}
	// Should prefer English title.
	if tender.Title != "Construction of municipal road" {
		t.Errorf("expected English title, got %q", tender.Title)
	}
	if tender.Description != "Road construction in Lisbon municipality" {
		t.Errorf("expected description, got %q", tender.Description)
	}
	if len(tender.CPVCodes) != 2 {
		t.Errorf("expected 2 CPV codes, got %d", len(tender.CPVCodes))
	}
	if tender.BuyerName != "Câmara Municipal de Lisboa" {
		t.Errorf("expected buyer name, got %q", tender.BuyerName)
	}
	if tender.BuyerCountry != "PRT" {
		t.Errorf("expected country PRT, got %q", tender.BuyerCountry)
	}
	if tender.Deadline == nil {
		t.Error("expected deadline to be parsed")
	}
	if tender.PublicationDate == nil {
		t.Error("expected publication_date to be parsed")
	}
	if tender.EstimatedValue == nil || *tender.EstimatedValue != 1500000 {
		t.Errorf("expected value 1500000, got %v", tender.EstimatedValue)
	}
	if tender.Currency != "EUR" {
		t.Errorf("expected currency EUR, got %q", tender.Currency)
	}
	if tender.NoticeURL != "https://ted.europa.eu/en/notice/-/detail/136388-2025" {
		t.Errorf("expected notice URL, got %q", tender.NoticeURL)
	}
}

func TestNormalize_MissingFields(t *testing.T) {
	notice := Notice{
		PublicationNumber: "999999-2025",
		TitleProc:         I18nText{"por": "Aviso mínimo"},
	}

	tender := Normalize(notice)

	if tender.SourceID != "999999-2025" {
		t.Errorf("expected source_id '999999-2025', got %q", tender.SourceID)
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
	if tender.NoticeURL == "" {
		t.Error("expected fallback notice URL")
	}
}

func TestExtractI18nText(t *testing.T) {
	tests := []struct {
		name string
		m    I18nText
		want string
	}{
		{"nil map", nil, ""},
		{"english preferred", I18nText{"por": "PT text", "eng": "EN text"}, "EN text"},
		{"portuguese fallback", I18nText{"por": "PT text", "fra": "FR text"}, "PT text"},
		{"any language fallback", I18nText{"deu": "DE text"}, "DE text"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractI18nText(tt.m)
			if got != tt.want {
				t.Errorf("extractI18nText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseTEDDate(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantNil bool
	}{
		{"empty", "", true},
		{"with timezone offset", "2025-03-03+01:00", false},
		{"with Z suffix", "2025-04-23Z", false},
		{"plain date", "2025-03-03", false},
		{"compact date", "20250303", false},
		{"invalid", "not-a-date", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTEDDate(tt.input)
			if (result == nil) != tt.wantNil {
				t.Errorf("parseTEDDate(%q) nil=%v, wantNil=%v", tt.input, result == nil, tt.wantNil)
			}
		})
	}
}
