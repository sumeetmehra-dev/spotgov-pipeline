package dados

// DatasetListResponse is the paginated response from dados.gov.pt datasets API.
type DatasetListResponse struct {
	Data     []Dataset `json:"data"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
	Total    int       `json:"total"`
}

// Dataset represents a dados.gov.pt dataset entry.
type Dataset struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	CreatedAt   string     `json:"created_at"`
	Resources   []Resource `json:"resources"`
	Organization *Organization `json:"organization"`
}

// Resource is a downloadable file within a dataset.
type Resource struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	URL    string `json:"url"`
	Format string `json:"format"`
}

// Organization represents the publishing entity.
type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ContractRecord represents a normalized contract from BASE/OCDS data.
type ContractRecord struct {
	ID                string   `json:"id"`
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	CPVCodes          []string `json:"cpv_codes"`
	BuyerName         string   `json:"buyer_name"`
	ContractValue     *float64 `json:"contract_value"`
	Currency          string   `json:"currency"`
	PublicationDate   string   `json:"publication_date"`
	Deadline          string   `json:"deadline"`
	ProcurementMethod string   `json:"procurement_method"`
}
