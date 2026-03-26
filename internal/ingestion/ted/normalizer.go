package ted

import (
	"encoding/json"
	"time"

	"github.com/lib/pq"
	"github.com/sumeetmehra/spotgov-pipeline/internal/model"
	"gorm.io/datatypes"
)

const (
	dateLayout     = "2006-01-02"
	datetimeLayout = "2006-01-02T15:04:05"
)

// Normalize converts a TED Notice into a unified Tender model.
func Normalize(n Notice) model.Tender {
	tender := model.Tender{
		Source:             "ted",
		SourceID:           n.PublicationNumber,
		Title:              n.Title,
		Description:        n.Description,
		CPVCodes:           pq.StringArray(n.CPVCodes),
		BuyerName:          n.BuyerName,
		BuyerCountry:       n.BuyerCountry,
		PlaceOfPerformance: n.PlaceOfPerformance,
		ProcurementMethod:  n.ProcurementMethod,
		EstimatedValue:     n.EstimatedValue,
		Currency:           defaultIfEmpty(n.Currency, "EUR"),
		NoticeURL:          n.NoticeURL,
	}

	if t := parseDate(n.Deadline); t != nil {
		tender.Deadline = t
	}
	if t := parseDate(n.PublicationDate); t != nil {
		tender.PublicationDate = t
	}

	if raw, err := json.Marshal(n); err == nil {
		tender.RawData = datatypes.JSON(raw)
	}

	return tender
}

func parseDate(s string) *time.Time {
	if s == "" {
		return nil
	}
	for _, layout := range []string{datetimeLayout, dateLayout} {
		if t, err := time.Parse(layout, s); err == nil {
			return &t
		}
	}
	return nil
}

func defaultIfEmpty(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
