package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sumeetmehra/spotgov-pipeline/internal/matching"
	"go.uber.org/zap"
)

type MatchHandler struct {
	matcher *matching.Matcher
	logger  *zap.Logger
}

func NewMatchHandler(matcher *matching.Matcher, logger *zap.Logger) *MatchHandler {
	return &MatchHandler{
		matcher: matcher,
		logger:  logger.Named("handler.match"),
	}
}

func (h *MatchHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Match)
	r.Post("/embeddings", h.GenerateEmbeddings)
	return r
}

type matchRequest struct {
	CompanyID string `json:"company_id"`
}

func (h *MatchHandler) Match(w http.ResponseWriter, r *http.Request) {
	var req matchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid company_id"})
		return
	}

	matches, err := h.matcher.MatchCompany(r.Context(), companyID)
	if err != nil {
		h.logger.Error("matching failed", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "matching failed"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "matching complete",
		"matches": len(matches),
	})
}

func (h *MatchHandler) GenerateEmbeddings(w http.ResponseWriter, r *http.Request) {
	count, err := h.matcher.GenerateTenderEmbeddings(r.Context(), 100)
	if err != nil {
		h.logger.Error("embedding generation failed", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "embedding generation failed"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "embeddings generated",
		"embedded": count,
	})
}
