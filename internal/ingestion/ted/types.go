package ted

// SearchRequest represents the TED API v3 search request body.
type SearchRequest struct {
	Query              string   `json:"query"`
	Fields             []string `json:"fields"`
	Limit              int      `json:"limit,omitempty"`
	Page               int      `json:"page,omitempty"`
	Scope              string   `json:"scope,omitempty"`
	PaginationMode     string   `json:"paginationMode,omitempty"`
	OnlyLatestVersions bool     `json:"onlyLatestVersions,omitempty"`
}

// SearchResponse is the top-level TED API v3 search response.
type SearchResponse struct {
	Notices            []Notice `json:"notices"`
	TotalNoticeCount   int      `json:"totalNoticeCount"`
	IterationNextToken *string  `json:"iterationNextToken"`
	TimedOut           bool     `json:"timedOut"`
}

// I18nText maps ISO 639-2 language codes to text values.
// Example: {"por": "Construção de estrada", "eng": "Road construction"}
type I18nText map[string]string

// I18nTextArray maps ISO 639-2 language codes to arrays of text values (one per lot).
// Example: {"por": ["Lot 1 description", "Lot 2 description"]}
type I18nTextArray map[string][]string

// NoticeLinks contains URLs to different representations of the notice.
type NoticeLinks struct {
	XML  map[string]string `json:"xml"`
	HTML map[string]string `json:"html"`
}

// Notice represents a single TED procurement notice from the v3 API.
// Text fields use i18n maps; value fields use string arrays (one per lot).
type Notice struct {
	PublicationNumber string       `json:"publication-number"`
	NoticeType        string       `json:"notice-type"`
	PublicationDate   string       `json:"publication-date"`
	TitleProc         I18nText     `json:"title-proc"`
	DescriptionLot    I18nTextArray `json:"description-lot"`
	DescriptionProc   I18nText     `json:"description-proc"`
	OrgNameBuyer      I18nTextArray `json:"organisation-name-buyer"`
	OrgCountryBuyer   []string     `json:"organisation-country-buyer"`
	EstimatedValueLot []string     `json:"estimated-value-lot"`
	EstimatedValueCur []string     `json:"estimated-value-cur-lot"`
	DeadlineDateLot   []string     `json:"deadline-receipt-tender-date-lot"`
	CPVCodes          []string     `json:"classification-cpv"`
	ContractNature    []string     `json:"contract-nature"`
	Links             NoticeLinks  `json:"links"`
}
