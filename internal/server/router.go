package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sumeetmehra/spotgov-pipeline/internal/handler"
	"github.com/sumeetmehra/spotgov-pipeline/internal/ingestion"
	"github.com/sumeetmehra/spotgov-pipeline/internal/matching"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func NewRouter(db *gorm.DB, orchestrator *ingestion.Orchestrator, matcher *matching.Matcher, logger *zap.Logger) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(handler.RecoveryMiddleware(logger))
	r.Use(handler.CORSMiddleware)
	r.Use(handler.RequestLogger(logger))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	health := handler.NewHealthHandler(db)
	tenders := handler.NewTenderHandler(db, orchestrator, logger)
	companies := handler.NewCompanyHandler(db, logger)
	matches := handler.NewMatchHandler(matcher, logger)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", health.Health)
		r.Get("/stats", health.Stats)
		r.Mount("/tenders", tenders.Routes())
		r.Mount("/companies", companies.Routes())
		r.Mount("/match", matches.Routes())
	})

	return r
}
