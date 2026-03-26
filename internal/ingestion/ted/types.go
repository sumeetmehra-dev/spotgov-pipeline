package ted

// SearchRequest represents the TED API search request body.
type SearchRequest struct {
	Query string `json:"query"`
	// Fields to return in results. If empty, returns all fields.
	Fields []string `json:"fields,omitempty"`
	// Maximum results per page (max 250).
	Limit int `json:"limit,omitempty"`
	// Pagination token from previous response.
	Page int `json:"page,omitempty"`
}

// SearchResponse is the top-level TED API search response.
type SearchResponse struct {
	Notices []Notice `json:"notices"`
	Total   int      `json:"total"`
	Page    int      `json:"page"`
	Pages   int      `json:"pages"`
}

// Notice represents a single TED procurement notice.
type Notice struct {
	PublicationNumber  string   `json:"publication-number"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	CPVCodes           []string `json:"cpv-codes"`
	BuyerName          string   `json:"buyer-name"`
	BuyerCountry       string   `json:"buyer-country"`
	PlaceOfPerformance string   `json:"place-of-performance"`
	ProcurementMethod  string   `json:"procurement-method"`
	EstimatedValue     *float64 `json:"estimated-value"`
	Currency           string   `json:"currency"`
	Deadline           string   `json:"deadline"`
	PublicationDate    string   `json:"publication-date"`
	NoticeType         string   `json:"notice-type"`
	NoticeURL          string   `json:"notice-url"`
}
