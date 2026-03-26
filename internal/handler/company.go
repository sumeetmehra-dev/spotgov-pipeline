package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/sumeetmehra/procura/internal/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CompanyHandler struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewCompanyHandler(db *gorm.DB, logger *zap.Logger) *CompanyHandler {
	return &CompanyHandler{
		db:     db,
		logger: logger.Named("handler.company"),
	}
}

func (h *CompanyHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.Create)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Get("/{id}/matches", h.GetMatches)
	return r
}

type companyRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	CPVCodes    []string `json:"cpv_codes"`
	Keywords    []string `json:"keywords"`
	Countries   []string `json:"countries"`
	MinValue    *float64 `json:"min_value"`
	MaxValue    *float64 `json:"max_value"`
}

func (h *CompanyHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req companyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.Name == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}

	company := model.Company{
		Name:        req.Name,
		Description: req.Description,
		CPVCodes:    pq.StringArray(req.CPVCodes),
		Keywords:    pq.StringArray(req.Keywords),
		Countries:   pq.StringArray(req.Countries),
		MinValue:    req.MinValue,
		MaxValue:    req.MaxValue,
	}

	if err := h.db.Create(&company).Error; err != nil {
		h.logger.Error("failed to create company", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create company"})
		return
	}

	respondJSON(w, http.StatusCreated, company)
}

func (h *CompanyHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid company ID"})
		return
	}

	var company model.Company
	if err := h.db.First(&company, "id = ?", id).Error; err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "company not found"})
		return
	}

	respondJSON(w, http.StatusOK, company)
}

func (h *CompanyHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid company ID"})
		return
	}

	var company model.Company
	if err := h.db.First(&company, "id = ?", id).Error; err != nil {
		respondJSON(w, http.StatusNotFound, map[string]string{"error": "company not found"})
		return
	}

	var req companyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.CPVCodes != nil {
		updates["cpv_codes"] = pq.StringArray(req.CPVCodes)
	}
	if req.Keywords != nil {
		updates["keywords"] = pq.StringArray(req.Keywords)
	}
	if req.Countries != nil {
		updates["countries"] = pq.StringArray(req.Countries)
	}
	if req.MinValue != nil {
		updates["min_value"] = req.MinValue
	}
	if req.MaxValue != nil {
		updates["max_value"] = req.MaxValue
	}

	if err := h.db.Model(&company).Updates(updates).Error; err != nil {
		h.logger.Error("failed to update company", zap.Error(err))
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to update company"})
		return
	}

	h.db.First(&company, "id = ?", id)
	respondJSON(w, http.StatusOK, company)
}

func (h *CompanyHandler) GetMatches(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid company ID"})
		return
	}

	limit := queryInt(r, "limit", 20)
	minScore := 0.0
	if ms := r.URL.Query().Get("min_score"); ms != "" {
		fmt.Sscanf(ms, "%f", &minScore)
	}

	var matches []model.Match
	h.db.Preload("Tender").
		Where("company_id = ? AND score >= ?", id, minScore).
		Order("score DESC").
		Limit(limit).
		Find(&matches)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":  matches,
		"count": len(matches),
	})
}
