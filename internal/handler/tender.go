package handler

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sumeetmehra/procura/internal/ingestion"
	"github.com/sumeetmehra/procura/internal/model"
	"github.com/sumeetmehra/procura/internal/search"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TenderHandler struct {
	db           *gorm.DB
	orchestrator *ingestion.Orchestrator
	es           *search.ESClient
	logger       *zap.Logger
}

func NewTenderHandler(db *gorm.DB, orchestrator *ingestion.Orchestrator, es *search.ESClient, logger *zap.Logger) *TenderHandler {
	return &TenderHandler{
		db:           db,
		orchestrator: orchestrator,
		es:           es,
		logger:       logger.Named("handler.tender"),
	}
}

func (h *TenderHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Get("/search", h.Search)
	r.Post("/ingest", h.Ingest)
	r.Get("/{id}", h.Get)
	return r
}

// List returns paginated tenders with optional filters.
func (h *TenderHandler) List(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 1)
	pageSize := queryInt(r, "page_size", 20)
	if pageSize > 100 {
		pageSize = 100
	}

	query := h.db.Model(&model.Tender{})

	// Apply filters
	if cpv := r.URL.Query().Get("cpv"); cpv != "" {
		query = query.Where("? = ANY(cpv_codes)", cpv)
	}
	if country := r.URL.Query().Get("country"); country != "" {
		query = query.Where("buyer_country = ?", country)
	}
	if minVal := r.URL.Query().Get("min_value"); minVal != "" {
		if v, err := strconv.ParseFloat(minVal, 64); err == nil {
			query = query.Where("estimated_value >= ?", v)
		}
	}
	if maxVal := r.URL.Query().Get("max_value"); maxVal != "" {
		if v, err := strconv.ParseFloat(maxVal, 64); err == nil {
			query = query.Where("estimated_value <= ?", v)
		}
	}
	if after := r.URL.Query().Get("deadline_after"); after != "" {
		if t, err := time.Parse("2006-01-02", after); err == nil {
			query = query.Where("deadline >= ?", t)
		}
	}
	if q := r.URL.Query().Get("q"); q != "" {
		query = query.Where("title ILIKE ? OR description ILIKE ?", "%"+q+"%", "%"+q+"%")
	}

	var total int64
	query.Count(&total)

	var tenders []model.Tender
	offset := (page - 1) * pageSize
	query.Order("publication_date DESC NULLS LAST").
		Offset(offset).Limit(pageSize).
		Find(&tenders)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":        tenders,
		"page":        page,
		"page_size":   pageSize,
		"total":       total,
		"total_pages": int(math.Ceil(float64(total) / float64(pageSize))),
	})
}

// Get returns a single tender by ID.
func (h *TenderHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	parsed, err := uuid.Parse(id)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid tender ID"})
		return
	}

	var tender model.Tender
	if err := h.db.First(&tender, "id = ?", parsed).Error; err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "tender not found"})
		return
	}

	respondJSON(w, http.StatusOK, tender)
}

// Search performs full-text search on tenders using Elasticsearch with DB fallback.
func (h *TenderHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "query parameter 'q' is required"})
		return
	}

	limit := queryInt(r, "limit", 20)
	if limit > 100 {
		limit = 100
	}

	// Try Elasticsearch first.
	if h.es != nil {
		filters := search.SearchFilters{
			CPVCode: r.URL.Query().Get("cpv"),
			Country: r.URL.Query().Get("country"),
		}
		results, err := h.es.Search(r.Context(), q, filters, limit)
		if err != nil {
			h.logger.Warn("ES search failed, falling back to DB", zap.Error(err))
		} else {
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"data":   results,
				"query":  q,
				"count":  len(results),
				"engine": "elasticsearch",
			})
			return
		}
	}

	// Fallback: DB ILIKE search.
	var tenders []model.Tender
	h.db.Where("title ILIKE ? OR description ILIKE ?", "%"+q+"%", "%"+q+"%").
		Order("publication_date DESC NULLS LAST").
		Limit(limit).
		Find(&tenders)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":   tenders,
		"query":  q,
		"count":  len(tenders),
		"engine": "database",
	})
}

// Ingest triggers a manual ingestion run.
func (h *TenderHandler) Ingest(w http.ResponseWriter, r *http.Request) {
	count, err := h.orchestrator.Run(r.Context())
	if err != nil {
		h.logger.Error("ingestion failed", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "ingestion failed"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "ingestion complete",
		"upserted": count,
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func queryInt(r *http.Request, key string, fallback int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}
