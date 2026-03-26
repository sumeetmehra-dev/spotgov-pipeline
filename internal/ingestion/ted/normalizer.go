package ted

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/sumeetmehra/spotgov-pipeline/internal/model"
	"gorm.io/datatypes"
)

// Normalize converts a TED v3 Notice into a unified Tender model.
func Normalize(n Notice) model.Tender {
	title := extractI18nText(n.TitleProc)
	description := extractI18nTextArrayJoined(n.DescriptionLot)
	if description == "" {
		description = extractI18nText(n.DescriptionProc)
	}

	tender := model.Tender{
		Source:            "ted",
		SourceID:          n.PublicationNumber,
		Title:             title,
		Description:       description,
		CPVCodes:          pq.StringArray(n.CPVCodes),
		BuyerName:         extractI18nTextArrayFirst(n.OrgNameBuyer),
		BuyerCountry:      firstOrEmpty(n.OrgCountryBuyer),
		ProcurementMethod: firstOrEmpty(n.ContractNature),
		Currency:          firstOrDefault(n.EstimatedValueCur, "EUR"),
		NoticeURL:         buildNoticeURL(n),
	}

	if v := parseValue(n.EstimatedValueLot); v != nil {
		tender.EstimatedValue = v
	}
	if t := parseTEDDate(firstOrEmpty(n.DeadlineDateLot)); t != nil {
		tender.Deadline = t
	}
	if t := parseTEDDate(n.PublicationDate); t != nil {
		tender.PublicationDate = t
	}

	if raw, err := json.Marshal(n); err == nil {
		tender.RawData = datatypes.JSON(raw)
	}

	return tender
}

// extractI18nText returns text preferring English, then Portuguese, then any language.
func extractI18nText(m I18nText) string {
	if m == nil {
		return ""
	}
	// Prefer English, then Portuguese, then first available.
	for _, lang := range []string{"eng", "por", "mul"} {
		if v, ok := m[lang]; ok && v != "" {
			return v
		}
	}
	for _, v := range m {
		if v != "" {
			return v
		}
	}
	return ""
}

// extractI18nTextArrayFirst returns the first element preferring English then Portuguese.
func extractI18nTextArrayFirst(m I18nTextArray) string {
	if m == nil {
		return ""
	}
	for _, lang := range []string{"eng", "por", "mul"} {
		if arr, ok := m[lang]; ok && len(arr) > 0 && arr[0] != "" {
			return arr[0]
		}
	}
	for _, arr := range m {
		if len(arr) > 0 && arr[0] != "" {
			return arr[0]
		}
	}
	return ""
}

// extractI18nTextArrayJoined joins all elements of the preferred language with newlines.
func extractI18nTextArrayJoined(m I18nTextArray) string {
	if m == nil {
		return ""
	}
	for _, lang := range []string{"eng", "por", "mul"} {
		if arr, ok := m[lang]; ok && len(arr) > 0 {
			return strings.Join(arr, "\n")
		}
	}
	for _, arr := range m {
		if len(arr) > 0 {
			return strings.Join(arr, "\n")
		}
	}
	return ""
}

// parseTEDDate parses TED date formats: "2025-03-03+01:00", "2025-04-23Z", "2025-03-03".
func parseTEDDate(s string) *time.Time {
	if s == "" {
		return nil
	}
	// Strip timezone suffixes to normalize.
	s = strings.TrimSuffix(s, "Z")
	if idx := strings.Index(s, "+"); idx > 0 {
		s = s[:idx]
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return &t
	}
	if t, err := time.Parse("20060102", s); err == nil {
		return &t
	}
	return nil
}

func parseValue(vals []string) *float64 {
	if len(vals) == 0 {
		return nil
	}
	v, err := strconv.ParseFloat(vals[0], 64)
	if err != nil {
		return nil
	}
	return &v
}

func firstOrEmpty(s []string) string {
	if len(s) == 0 {
		return ""
	}
	return s[0]
}

func firstOrDefault(s []string, fallback string) string {
	if len(s) == 0 || s[0] == "" {
		return fallback
	}
	return s[0]
}

func buildNoticeURL(n Notice) string {
	if n.Links.HTML != nil {
		for _, lang := range []string{"ENG", "MUL"} {
			if url, ok := n.Links.HTML[lang]; ok {
				return url
			}
		}
		for _, url := range n.Links.HTML {
			return url
		}
	}
	if n.PublicationNumber != "" {
		return "https://ted.europa.eu/en/notice/-/detail/" + n.PublicationNumber
	}
	return ""
}
