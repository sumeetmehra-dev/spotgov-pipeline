package handler

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	status := "healthy"
	dbStatus := "up"

	sqlDB, err := h.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		status = "degraded"
		dbStatus = "down"
	}

	w.Header().Set("Content-Type", "application/json")
	if status != "healthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(map[string]string{
		"status":   status,
		"database": dbStatus,
	})
}

func (h *HealthHandler) Stats(w http.ResponseWriter, r *http.Request) {
	var totalTenders int64
	var tedCount int64
	var dadosCount int64

	h.db.Model(&struct{}{}).Table("tenders").Count(&totalTenders)
	h.db.Model(&struct{}{}).Table("tenders").Where("source = ?", "ted").Count(&tedCount)
	h.db.Model(&struct{}{}).Table("tenders").Where("source = ?", "dados").Count(&dadosCount)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_tenders": totalTenders,
		"by_source": map[string]int64{
			"ted":   tedCount,
			"dados": dadosCount,
		},
	})
}
