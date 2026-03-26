package ted

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/sumeetmehra/procura/internal/model"
	"gorm.io/datatypes"
)

// Matches lot labels like "Lote 1", "Lot 2 - BrandName", "Lote n.º 3", "LOTE 1 - Trojan"
var lotLabelPattern = regexp.MustCompile(`(?i)^(lote|lot|partie|los|partida)\s+(n\.?º?\s*)?\d+(\s*[-–:].+)?$`)

// Normalize converts a TED v3 Notice into a unified Tender model.
func Normalize(n Notice) model.Tender {
	title := extractI18nText(n.TitleProc)
	description := buildDescription(n, title)

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

// buildDescription extracts the best description from TED notice fields.
// Prefers description-proc (overall procurement description) over description-lot (per-lot labels).
// Filters out lot labels like "Lote 1", "Lote 2" that aren't real descriptions.
// Deduplicates against the title to avoid showing the same text twice.
func buildDescription(n Notice, title string) string {
	procDesc := extractI18nText(n.DescriptionProc)
	lotDesc := extractSubstantiveLotText(n.DescriptionLot)

	// Prefer proc-level description (usually richer).
	if procDesc != "" && !strings.EqualFold(procDesc, title) {
		if lotDesc != "" && lotDesc != procDesc {
			return procDesc + "\n\n" + lotDesc
		}
		return procDesc
	}

	// Fall back to lot descriptions.
	if lotDesc != "" && !strings.EqualFold(lotDesc, title) {
		return lotDesc
	}

	// Last resort: return whichever is non-empty.
	// If both match the title, return empty — duplicating the title adds no value.
	if procDesc != "" && !strings.EqualFold(procDesc, title) {
		return procDesc
	}
	if lotDesc != "" && !strings.EqualFold(lotDesc, title) {
		return lotDesc
	}
	return ""
}

// extractSubstantiveLotText joins lot descriptions after filtering out bare lot labels.
func extractSubstantiveLotText(m I18nTextArray) string {
	if m == nil {
		return ""
	}
	for _, lang := range []string{"eng", "por", "mul"} {
		if arr, ok := m[lang]; ok && len(arr) > 0 {
			filtered := filterSubstantive(arr)
			if len(filtered) > 0 {
				return strings.Join(filtered, "\n")
			}
		}
	}
	for _, arr := range m {
		if len(arr) > 0 {
			filtered := filterSubstantive(arr)
			if len(filtered) > 0 {
				return strings.Join(filtered, "\n")
			}
		}
	}
	return ""
}

// filterSubstantive removes entries that are just lot labels (e.g. "Lote 1") or too short to be useful.
func filterSubstantive(entries []string) []string {
	var result []string
	for _, s := range entries {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if lotLabelPattern.MatchString(s) {
			continue
		}
		result = append(result, s)
	}
	return result
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
