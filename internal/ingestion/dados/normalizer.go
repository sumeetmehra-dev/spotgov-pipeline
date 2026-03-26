package dados

import (
	"encoding/json"
	"time"

	"github.com/sumeetmehra/spotgov-pipeline/internal/model"
	"gorm.io/datatypes"
)

// Normalize converts a dados.gov.pt Dataset into a unified Tender model.
func Normalize(d Dataset) model.Tender {
	tender := model.Tender{
		Source:      "dados",
		SourceID:    "dados-" + d.ID,
		Title:       d.Title,
		Description: d.Description,
		Currency:    "EUR",
	}

	if d.Organization != nil {
		tender.BuyerName = d.Organization.Name
	}
	tender.BuyerCountry = "PT"

	if t := parseDate(d.CreatedAt); t != nil {
		tender.PublicationDate = t
	}

	if raw, err := json.Marshal(d); err == nil {
		tender.RawData = datatypes.JSON(raw)
	}

	return tender
}

func parseDate(s string) *time.Time {
	if s == "" {
		return nil
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return &t
		}
	}
	return nil
}
